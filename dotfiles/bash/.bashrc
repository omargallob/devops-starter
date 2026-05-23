# devops-starter bash configuration

# Resolve real dotfiles directory (works through symlinks)
DOTFILES_DIR="$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")/../shared"
[ -d "$DOTFILES_DIR" ] && {
    source "$DOTFILES_DIR/env.sh"
    source "$DOTFILES_DIR/aliases.sh"
    source "$DOTFILES_DIR/functions.sh"
}

# History
HISTSIZE=50000
HISTFILESIZE=50000
HISTCONTROL=ignoreboth:erasedups
shopt -s histappend
shopt -s cdspell
shopt -s globstar

# PS1 fallback
PS1='\[\033[01;32m\]\u@\h\[\033[00m\]:\[\033[01;34m\]\w\[\033[00m\]\$ '

# Tool initialization
command -v starship >/dev/null && eval "$(starship init bash)"
command -v zoxide >/dev/null && eval "$(zoxide init bash)"
command -v direnv >/dev/null && eval "$(direnv hook bash)"
command -v atuin >/dev/null && eval "$(atuin init bash)"
command -v mise >/dev/null && eval "$(mise activate bash)"
