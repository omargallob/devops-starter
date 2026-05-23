# devops-starter zsh configuration

# Resolve real dotfiles directory (works through symlinks)
DOTFILES_DIR="$(dirname "$(readlink -f "${(%):-%x}")")/../shared"
[ -d "$DOTFILES_DIR" ] && {
    source "$DOTFILES_DIR/env.sh"
    source "$DOTFILES_DIR/aliases.sh"
    source "$DOTFILES_DIR/functions.sh"
}

# History
HISTSIZE=50000
SAVEHIST=50000
HISTFILE=~/.zsh_history
setopt appendhistory
setopt sharehistory
setopt hist_ignore_dups
setopt hist_ignore_space

# Completion
autoload -Uz compinit && compinit

# Key bindings - history search
bindkey '^[[A' history-search-backward
bindkey '^[[B' history-search-forward

# Tool initialization
command -v starship >/dev/null && eval "$(starship init zsh)"
command -v zoxide >/dev/null && eval "$(zoxide init zsh)"
command -v direnv >/dev/null && eval "$(direnv hook zsh)"
command -v atuin >/dev/null && eval "$(atuin init zsh)"
command -v mise >/dev/null && eval "$(mise activate zsh)"
