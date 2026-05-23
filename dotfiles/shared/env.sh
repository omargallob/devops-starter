#!/usr/bin/env sh
# Shared environment variables for devops-starter

export EDITOR=nvim
export VISUAL=nvim

export LANG=en_US.UTF-8

# XDG Base Directory
export XDG_CONFIG_HOME="$HOME/.config"
export XDG_DATA_HOME="$HOME/.local/share"
export XDG_CACHE_HOME="$HOME/.cache"

# Go
export GOPATH="$HOME/go"

# PATH
export PATH="$HOME/.local/bin:$GOPATH/bin:$HOME/.cargo/bin:$HOME/.mise/shims:$PATH"
