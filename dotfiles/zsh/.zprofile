# devops-starter zsh login profile

# Source env if not already loaded
if [ -z "$DOTFILES_ENV_LOADED" ]; then
    DOTFILES_DIR="$(dirname "$(readlink -f "${(%):-%x}")")/../shared"
    [ -f "$DOTFILES_DIR/env.sh" ] && source "$DOTFILES_DIR/env.sh"
    export DOTFILES_ENV_LOADED=1
fi

umask 022
