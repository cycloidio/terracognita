#!/bin/bash

TC_DIR=$PWD
TMP_DIR=/tmp

update() {
  PROVIDER_LIST=$1

  if [ "$PROVIDER_LIST" = "all" ]; then
    PROVIDER_LIST="aws azurerm"
  fi

  for PROVIDER in $PROVIDER_LIST; do
    echo $PROVIDER
    update_terraform_provider
    update_terracognita
  done
}

update_terraform_provider() {
  PROVIDER_DIR=${TMP_DIR}/terraform-provider-${PROVIDER}

  if [ ! -d $PROVIDER_DIR ]; then
    echo "Clonning terraform-provider-${PROVIDER} ..."
    git clone git@github.com:cycloidio/terraform-provider-${PROVIDER}.git ${PROVIDER_DIR}
  fi

  cd ${PROVIDER_DIR}
  git remote add upstream https://github.com/hashicorp/terraform-provider-${PROVIDER}.git
  git fetch upstream
  TAG=$(git tag | tail -n1)
  git checkout cycloid
  git rebase --onto ${TAG} upstream/main
  git push -f origin cycloid
  LASTCOMMIT=$(git rev-parse --short HEAD)
}

update_terracognita() {
  echo "Update Terracognita go.mod with ${PROVIDER} ${TAG}"
  cd $TC_DIR
  go mod edit -replace github.com/hashicorp/terraform-provider-${PROVIDER}=github.com/cycloidio/terraform-provider-${PROVIDER}@$LASTCOMMIT
  go mod tidy

  echo "Update README.md ..."

  if [ "$PROVIDER" = "aws" ]; then
      sed -i "s/^ \* AWS: .*/ * AWS: $TAG/" README.md
  elif [ "$PROVIDER" = "azurerm" ]; then
      sed -i "s/^ \* AzureRM: .*/ * AzureRM: $TAG/" README.md
  fi

  echo "${PROVIDER} updated to ${TAG}"
}

case "$1" in
    'aws')
        update aws
        ;;
    'azurerm')
        update azurerm
        ;;
    'all')
        update all
        ;;
    *)
        echo "Usage: $0 {aws|azurerm|all}"
        exit 1
        ;;
esac

exit 0
