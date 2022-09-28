#! /usr/bin/env bash

# Workstation Setup
pushd ~/workspace/
    rm -rf workstation-setup
    git clone https://github.com/pivotal/workstation-setup
    pushd workstation-setup
      ./setup.sh c golang docker
    popd
popd

# Install with brew:
brew cask install google-cloud-sdk
brew install neovim bash bash-completion gnu-sed kubernetes-cli helm minikube

# Setup newer bash version as shell:
<<<"/usr/local/bin/bash" sudo tee -a /etc/shells >/dev/null
chsh -s /usr/local/bin/bash

# Disable accent suggestions on long press
defaults write -g ApplePressAndHoldEnabled -bool false

# git global settings
git config --global commit.verbose true
git config --global rebase.autoSquash true
git config --global rebase.autoStash true

# intra-line diff higlighting
cat >~/bin/diff-highlight <<_EOF
#!/bin/sh

# /usr/local/Cellar/git/<version>/libexec/git-core
gitbindir=$(dirname $(which git))
# /usr/local/Cellar/git/<version>/share/git-core/contrib
gitcontribdir=${gitbindir/\/libexec\///share/}/contrib
# /usr/local/Cellar/git/<version>/share/git-core/contrib/diff-highlight/diff-highlight
diffhighlight=${gitcontribdir}/diff-highlight/diff-highlight

exec $diffhighlight
_EOF
chmod +x ~/bin/diff-highlight
git config --global pager.log  'diff-highlight | less'
git config --global pager.show 'diff-highlight | less'
git config --global pager.diff 'diff-highlight | less'
git config --global interactive.diffFilter 'diff-highlight'

# Setup bash theme
mkdir -p ~/.bash_it/themes/gp4k
ln -sf ~/workspace/greenplum-for-kubernetes/workstation-setup/gp4k.theme.bash ~/.bash_it/themes/gp4k/gp4k.theme.bash

# Setup bash_profile
gsed -i 's#export BASH_IT_THEME=.*#export BASH_IT_THEME=\"gp4k\"#g' ~/.bash_profile

echo '[ -f ~/.bashrc ] && source ~/.bashrc' >> ~/.bash_profile

# Setup bashrc
ln -sf ~/workspace/greenplum-for-kubernetes/workstation-setup/bashrc ~/.bashrc
