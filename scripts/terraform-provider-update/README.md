# terraform-provider-update

## update.sh

Script used by `make update-terraform-provider`.
The goal is to help updating Terraform and Terraform providers used by Terracognita.


### Actions

Here is a description of actions done

update_terraform_provider
 * Git clone the Cycloid Provider fork repository
 * Update the code from official remote upstream
 * Create a cycloid branch matching the release version
 * Move code out of internal directory
 * Change the go import to match the new path
 * Apply specific `code_fix_*` script (if needed)
 * Git push a new commit/tags

update_terracognita
 * Update Terracognita Go mod with the latest
 * Update README.md
 * Apply specific `terracognita_fix_*` script (if needed)
