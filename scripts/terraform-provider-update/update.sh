#!/bin/bash

TC_DIR=$PWD
TMP_DIR=/tmp

update() {
  PROVIDER_LIST=$1

  if [ "$PROVIDER_LIST" = "all" ]; then
    PROVIDER_LIST="aws azurerm google"
  fi

  for PROVIDER in $PROVIDER_LIST; do
    echo $PROVIDER
    update_terraform_provider
    update_terracognita
  done
}

# Clone the provider code to update it and get the tag/commit
update_terraform_provider() {
  PROVIDER_DIR=${TMP_DIR}/terraform-provider-${PROVIDER}
  GIT_ORG=cycloidio

  # google do not have lib under internal and do not need a fork. Use the upstream repo
  if [ "$PROVIDER_LIST" = "google" ]; then
    GIT_ORG=hashicorp
  fi

  # azure and aws moved the lib under internal. We use our fork to fix it
  if [ ! -d $PROVIDER_DIR ]; then
    echo "Clonning terraform-provider-${PROVIDER} ..."
    git clone git@github.com:${GIT_ORG}/terraform-provider-${PROVIDER}.git ${PROVIDER_DIR}
  fi

  cd ${PROVIDER_DIR}
  if [ "$PROVIDER_LIST" = "google" ]; then
    git fetch
    TAG=$(git tag | tail -n1)
    git checkout ${TAG}
    LASTCOMMIT=$(git rev-parse --short HEAD)
  else
    git remote add upstream https://github.com/hashicorp/terraform-provider-${PROVIDER}.git
    git fetch upstream
    TAG=$(git tag | tail -n1)
    git checkout cycloid
    git rebase --onto ${TAG} upstream/main
    git push -f origin cycloid
    LASTCOMMIT=$(git rev-parse --short HEAD)
  fi
}

update_terracognita() {
  GIT_ORG=cycloidio
  # google do not have lib under internal and do not need a fork. Use the upstream repo
  if [ "$PROVIDER_LIST" = "google" ]; then
    GIT_ORG=hashicorp
  fi

  echo "Update Terracognita go.mod with ${PROVIDER} ${TAG}"
  cd $TC_DIR
  go mod edit -replace github.com/hashicorp/terraform-provider-${PROVIDER}=github.com/${GIT_ORG}/terraform-provider-${PROVIDER}@$LASTCOMMIT
  go mod tidy

  echo "Update README.md ..."
  if [ "$PROVIDER" = "aws" ]; then
      sed -i "s/^ \* AWS: .*/ * AWS: $TAG/" README.md
  elif [ "$PROVIDER" = "azurerm" ]; then
      sed -i "s/^ \* AzureRM: .*/ * AzureRM: $TAG/" README.md
  elif [ "$PROVIDER" = "google" ]; then
      sed -i "s/^ \* Google: .*/ * Google: $TAG/" README.md
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
    all)
        update all
        ;;
    *)
        echo "Usage: $0 {aws|azurerm|google|all}"
        exit 1
        ;;
esac

exit 0
