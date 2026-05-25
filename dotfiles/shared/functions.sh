#!/usr/bin/env sh
# Shared shell functions

# Create directory and cd into it
mkcd() {
	mkdir -p "$1" || return 1
	cd "$1" || return 1
}

# Extract any archive format
extract() {
	if [ -f "$1" ]; then
		case "$1" in
		*.tar.bz2) tar xjf "$1" ;;
		*.tar.gz) tar xzf "$1" ;;
		*.tar.xz) tar xJf "$1" ;;
		*.bz2) bunzip2 "$1" ;;
		*.gz) gunzip "$1" ;;
		*.tar) tar xf "$1" ;;
		*.tbz2) tar xjf "$1" ;;
		*.tgz) tar xzf "$1" ;;
		*.zip) unzip "$1" ;;
		*.Z) uncompress "$1" ;;
		*.7z) 7z x "$1" ;;
		*.zst) unzstd "$1" ;;
		*) echo "Cannot extract '$1'" ;;
		esac
	else
		echo "'$1' is not a valid file"
	fi
}

# Show process on a given port
port() {
	lsof -i :"$1"
}

# Kubectl logs with follow
klog() {
	kubectl logs -f "$@"
}

# Quick weather
weather() {
	curl -s "wttr.in/${1:-}"
}
