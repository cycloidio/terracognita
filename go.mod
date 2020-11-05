module github.com/cycloidio/terracognita

require (
	github.com/Azure/azure-sdk-for-go v46.4.0+incompatible
	github.com/Azure/go-autorest/autorest v0.11.10
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/aws/aws-sdk-go v1.35.9
	github.com/chr4/pwgen v1.1.0
	github.com/cycloidio/tfdocs v0.0.0-20200111145532-e6a80a93d7cc
	github.com/go-kit/kit v0.9.0
	github.com/golang/mock v1.4.4
	github.com/hashicorp/go-azure-helpers v0.12.0
	github.com/hashicorp/hcl2 v0.0.0-20190821123243-0c888d1241f6
	github.com/hashicorp/hil v0.0.0-20190212132231-97b3a9cdfa93 // indirect
	github.com/hashicorp/terraform v0.12.26
	github.com/hashicorp/terraform-plugin-sdk v1.16.0
	github.com/hashicorp/terraform-provider-google v1.20.1-0.20200922000057-78da461b151a
	github.com/hashicorp/vault v1.0.3 // indirect
	github.com/jinzhu/inflection v1.0.0
	github.com/keybase/go-crypto v0.0.0-20181127160227-255a5089e85a // indirect
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.6.1
	github.com/terraform-providers/terraform-provider-aws v1.60.1-0.20200807230610-d5346d47e3af
	github.com/terraform-providers/terraform-provider-azurerm v1.44.1-0.20201029183808-d721bcc1bb55
	github.com/terraform-providers/terraform-provider-random v2.0.0+incompatible // indirect
	github.com/zclconf/go-cty v1.5.1
	google.golang.org/api v0.31.1-0.20200914161323-7b3b1fe2dc94
)

// Force an specific version if not the AWS provider does not compile
replace github.com/hashicorp/aws-sdk-go-base v0.6.0 => github.com/hashicorp/aws-sdk-go-base v0.5.0

// To remove the panic issue of using TF
replace github.com/hashicorp/terraform => github.com/cycloidio/terraform v0.13.5-cy

go 1.14
