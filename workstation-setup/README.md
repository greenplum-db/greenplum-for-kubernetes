Run setup.sh to get started. This will run setup.sh in https://github.com/pivotal/workstation-setup with "c golang docker".

# goland plugins to install
1. makefile support
2. bashsupport
3. jsonnet

# Install goimports and go fmt plugin
1. Go to preferences in goland
2. Search for file watchers
3. Add new and select goimports. This might prompt you to install the goimports plugin.
4. Default settings are fine
6. Also add go fmt watcher the same way as goimports

# Set up git author
https://github.com/pivotal/git-author
1. brew install pivotal/tap/git-together
1. brew install pivotal/tap/git-author
1. git config --global --add include.path ~/.git-together
1. git config --file ~/.git-together --add git-together.domain vmware.com
1. git config --file ~/.git-together --add git-together.authors.<<initials>> '<<Full Name>>; <<email alias>>'
1. git clone git@github.com:pivotal/git-author and run ./git-author/setup.sh

# How to commit
1. STORY_NUM="#123456 brief description" git author nk kh
2. git commit
