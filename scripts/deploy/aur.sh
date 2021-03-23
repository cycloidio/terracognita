#!/usr/bin/sh

# this script can be used manually or in a CI context in order to push
# terracognita archlinux package to the AUR
# usage:
# ```
# export SSH_PRIVATE_KEY=$(cat /path/to/aur/ssh/key)
# sh ./deploy.sh
# ```

# script requirements:
# * curl
# * jq
# * wget
# * git

function error {
    echo $1
    exit 1
}

set -xe

# SSH_PRIVATE_KEY is the private key used to push / pull on the AUR package
[[ -z "$SSH_PRIVATE_KEY" ]] && error "SSH_PRIVATE_KEY must be set"

# fetch the latest tag name
tag_name=$(curl --silent "https://api.github.com/repos/cycloidio/terracognita/releases/latest" |
	jq -r .tag_name
)

source_i386="https://github.com/cycloidio/terracognita/releases/download/$tag_name/terracognita-linux-386.tar.gz"
source_amd64="https://github.com/cycloidio/terracognita/releases/download/$tag_name/terracognita-linux-amd64.tar.gz"

# download the two assets for linux (amd64, i386)
wget $source_i386
wget $source_amd64

# compute sha256 sums
sha256sum_i386=$(sha256sum terracognita-linux-386.tar.gz | awk '{ print $1 }')
sha256sum_amd64=$(sha256sum terracognita-linux-amd64.tar.gz | awk '{ print $1 }')

# save the SSH_PRIVATE_KEY in a file, to be used as identity
echo "$SSH_PRIVATE_KEY" > id_rsa.aur; chmod 600 ./id_rsa.aur

# clone the actual AUR
GIT_SSH_COMMAND="ssh -o StrictHostKeyChecking=no -i ./id_rsa.aur" git clone ssh://aur@aur.archlinux.org/terracognita.git && cd terracognita

# update the manifests (PKGBUILD and .SRCINFO)
# update the package version, the URLs and the sha256sums
# ${tag_name:1} to prevent adding the 'v' character of the version
sed -i "s/pkgver=[^\"]*/pkgver=${tag_name:1}/" PKGBUILD
sed -i "s/sha256sums_i386=[^\"]*/sha256sums_i386=('$sha256sum_i386')/" PKGBUILD
sed -i "s/sha256sums_x86_64=[^\"]*/sha256sums_x86_64=('$sha256sum_amd64')/" PKGBUILD

sed -i "s|source_i386 = [^\"]*|source_i386 = $source_i386|" .SRCINFO
sed -i "s/sha256sums_i386 = [^\"]*/sha256sums_i386 = $sha256sum_i386/" .SRCINFO
sed -i "s|source_x86_64 = [^\"]*|source_x86_64 = $source_amd64|" .SRCINFO
sed -i "s/sha256sums_x86_64 = [^\"]*/sha256sums_x86_64 = $sha256sum_amd64/" .SRCINFO
sed -i "s/pkgver = [^\"]*/pkgver = ${tag_name:1}/" .SRCINFO

# show git diff
git --no-pager diff

# configure git
git config --global user.name "cycloid"
git config --global user.email "cycloid@build"

# commit the files
# TODO: sign the commit by the CI
git commit -am "terracognita: bump to version $tag_name"

# push the files
GIT_SSH_COMMAND="ssh -o StrictHostKeyChecking=no -i ../id_rsa.aur" git push -u origin master

cd ../; rm -rf terracognita* id_rsa.aur
