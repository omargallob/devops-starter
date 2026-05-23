#!/usr/bin/env sh
# Shared aliases - rust-based tool replacements

# Modern coreutils (rust)
alias ls='eza'
alias ll='eza -la'
alias la='eza -la'
alias lt='eza --tree'
alias cat='bat'
alias grep='rg'
alias find='fd'

# Editor
alias vim='nvim'
alias vi='nvim'

# Kubernetes
alias k='kubectl'
alias kx='kubectx'
alias kn='kubens'

# Infrastructure
alias tf='terraform'
alias tg='tofu'

# Git
alias g='git'
alias gs='git status'
alias gd='git diff'
alias gl='git log --oneline'

# Docker
alias dc='docker compose'

# TUI
alias lg='lazygit'

# Navigation
alias ..='cd ..'
alias ...='cd ../..'
