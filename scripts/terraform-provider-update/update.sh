#!/bin/bash

TC_DIR=$PWD
TMP_DIR=/tmp

update() {
  PROVIDER_LIST=$1

  if [ "$PROVIDER_LIST" = "all" ]; then
    PROVIDER_LIST="aws azurerm google vsphere terraform"
  fi

  for PROVIDER in $PROVIDER_LIST; do
    echo "### Update: ${PROVIDER}"
    update_terraform_provider
    update_terracognita
  done
}

code_fix_terraform() {
  # Fix rpc import (linked to https://github.com/hashicorp/terraform-plugin-sdk/issues/268#issuecomment-562667048)
  sed -i 's/rpcFriendlyDiag/rpcFriendlyDiagTF/' tfdiags/*.go
}

move_out_internal() {
  # Fetch code upstream
  git remote add upstream https://github.com/hashicorp/${GIT_REPO}.git
  git fetch upstream
  # grep -v -: do not get alpha/beta/rc tags
  if [[ -z "${TAG}" ]]; then
    TAG=$(git tag --sort=-creatordate | grep -v '-' | head -n1 )
  fi
  # create a new branch for the release
  git branch cy-${TAG}
  git checkout cy-${TAG}
  # Use the latest tag release
  git reset --hard ${TAG}
  git clean -fdx

  # move everything out of internal
  for dir in $(ls internal)
  do
    mv internal/${dir} .
    find . -type f -name "*.go" -print0 | xargs -0 sed -i "s@/internal/${dir}@/${dir}@"
  done

  # Apply extra fixes
  if [ "$PROVIDER" = "terraform" ]; then
    code_fix_terraform
  fi

  # commit the fixes
  git add .
  git commit -m "cycloid fixes"

  # create a tag to help following releases
  git tag -f ${TAG}-cy
  git push -f origin ${TAG}-cy

  # Push the updated code to cycloid repo
  git push -f origin cy-${TAG}
  # Get the commit id to use in go.mod
  LASTCOMMIT=$(git rev-parse --short HEAD)
}

# Clone the provider code to update it and get the tag/commit
update_terraform_provider() {
  PROVIDER_LOCAL_DIR=${TMP_DIR}/terraform-provider-${PROVIDER}
  GIT_ORG=cycloidio
  GIT_REPO=terraform-provider-${PROVIDER}

  # google|vsphere: do not have lib under internal and do not need a fork. Use the upstream repo
  if [ "$PROVIDER" = "google" ] || [ "$PROVIDER" = "vsphere" ]; then
    GIT_ORG=hashicorp
  fi

  # terraform: change to match the terraform repository name
  if [ "$PROVIDER" = "terraform" ]; then
    GIT_REPO=terraform
  fi

  # azure and aws moved the lib under internal. We use our fork to fix it
  if [ ! -d $PROVIDER_LOCAL_DIR ]; then
    echo "Clonning terraform-provider-${PROVIDER} ..."
    git clone git@github.com:${GIT_ORG}/${GIT_REPO}.git ${PROVIDER_LOCAL_DIR}
  fi

  cd ${PROVIDER_LOCAL_DIR}
  # Use the upstream version
  if [ "$PROVIDER" = "google" ] || [ "$PROVIDER" = "vsphere" ]; then
    git fetch
    if [[ -z "${TAG}" ]]; then
      TAG=$(git tag --sort=-creatordate | grep -v '-' | head -n1 )
    fi
    git checkout ${TAG}
    LASTCOMMIT=$(git rev-parse --short HEAD)
  else
    # update our fork by moving out code under internal directory
    move_out_internal
  fi
}

terracognita_fix_terraform() {
  # get the version without v
  prefix="v"
  terraform_version=${TAG#"$prefix"}
  # Fix unit test with terraform version
  sed -i -E "s/terraform_version\": \"[0-9\.]+\"/terraform_version\": \"${terraform_version}\"/" state/writer_test.go
}

update_terracognita() {
  GIT_ORG=cycloidio
  GIT_REPO=terraform-provider-${PROVIDER}

  # google|vsphere: do not have lib under internal and do not need a fork. Use the upstream repo
  if [ "$PROVIDER" = "google" ] || [ "$PROVIDER" = "vsphere" ]; then
    GIT_ORG=hashicorp
  fi
  # terraform: change to match the terraform repository name
  if [ "$PROVIDER" = "terraform" ]; then
    GIT_REPO=terraform
  fi

  echo "Update Terracognita go.mod with ${PROVIDER} ${TAG}"
  cd $TC_DIR
  go mod edit -replace github.com/hashicorp/${GIT_REPO}=github.com/${GIT_ORG}/${GIT_REPO}@$LASTCOMMIT
  go mod tidy

  prefix="v"
  provider_version=${TAG#"$prefix"}
  echo "Update README.md ..."
  if [ "$PROVIDER" = "aws" ]; then
      sed -i "s/^ \* AWS: .*/ * AWS: $TAG/" README.md
      sed -i "s/^const version = .*/const version = \"${provider_version}\"/" aws/provider.go
  elif [ "$PROVIDER" = "azurerm" ]; then
      sed -i "s/^ \* AzureRM: .*/ * AzureRM: $TAG/" README.md
      sed -i "s/^const version = .*/const version = \"${provider_version}\"/" azurerm/provider.go
  elif [ "$PROVIDER" = "google" ]; then
      sed -i "s/^ \* Google: .*/ * Google: $TAG/" README.md
      sed -i "s/^const version = .*/const version = \"${provider_version}\"/" google/provider.go
  elif [ "$PROVIDER" = "vsphere" ]; then
      sed -i "s/^ \* vSphere: .*/ * vSphere: $TAG/" README.md
      sed -i "s/^const version = .*/const version = \"${provider_version}\"/" vsphere/provider.go
  elif [ "$PROVIDER" = "terraform" ]; then
      sed -E -i "s/Terraform \([0-9\.]+\)/Terraform ($TAG)/" README.md
      terracognita_fix_terraform
  fi

  echo "${PROVIDER} updated to ${TAG}"
}

case "$1" in
    aws)
        update aws
        ;;
    azurerm)
        update azurerm
        ;;
    gcp|google)
        update google
        ;;
    vsphere)
        update vsphere
        ;;
    terraform)
        update terraform
        ;;
    all)
        update all
        ;;
    *)
        echo "Usage: $0 {aws|azurerm|google|vsphere|terraform|all}"
        exit 1
        ;;
esac

exit 0
