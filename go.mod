module github.com/cycloidio/terracognita

require (
	github.com/Azure/azure-sdk-for-go v42.1.0+incompatible
	github.com/Azure/go-autorest/autorest v0.10.0
	github.com/aws/aws-sdk-go v1.31.9
	github.com/chr4/pwgen v1.1.0
	github.com/cycloidio/tfdocs v0.0.0-20200111145532-e6a80a93d7cc
	github.com/go-kit/kit v0.9.0
	github.com/golang/mock v1.4.3
	github.com/hashicorp/go-azure-helpers v0.10.0
	github.com/hashicorp/hcl v1.0.0
	github.com/hashicorp/hil v0.0.0-20190212132231-97b3a9cdfa93 // indirect
	github.com/hashicorp/terraform v0.12.26
	github.com/hashicorp/terraform-plugin-sdk v1.13.1
	github.com/hashicorp/vault v1.0.3 // indirect
	github.com/jinzhu/inflection v1.0.0
	github.com/keybase/go-crypto v0.0.0-20181127160227-255a5089e85a // indirect
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.6.1
	github.com/stretchr/testify v1.5.1
	github.com/terraform-providers/terraform-provider-aws v1.60.1-0.20200605194953-d15a5fede10a
	github.com/terraform-providers/terraform-provider-azurerm v1.44.1-0.20200605204846-1b0da11661ac
	github.com/terraform-providers/terraform-provider-google v1.20.1-0.20200605200304-d3a169b0353a
	github.com/terraform-providers/terraform-provider-random v2.0.0+incompatible // indirect
	github.com/zclconf/go-cty v1.4.2
	golang.org/x/tools v0.0.0-20200515220128-d3bf790afa53 // indirect
	google.golang.org/api v0.25.0
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999

replace github.com/terraform-providers/terraform-provider-tls v2.1.0+incompatible => github.com/terraform-providers/terraform-provider-tls v1.2.1-0.20190816230231-0790c4b40281

// This are directly from the https://github.com/terraform-providers/terraform-provider-azurerm/blob/master/go.mod
replace github.com/Azure/go-autorest => github.com/tombuildsstuff/go-autorest v14.0.1-0.20200416184303-d4e299a3c04a+incompatible

replace github.com/Azure/go-autorest/autorest => github.com/tombuildsstuff/go-autorest/autorest v0.10.1-0.20200416184303-d4e299a3c04a

replace github.com/Azure/go-autorest/autorest/azure/auth => github.com/tombuildsstuff/go-autorest/autorest/azure/auth v0.4.3-0.20200416184303-d4e299a3c04a

// To remove the panic issue of using TF
replace github.com/hashicorp/terraform => github.com/cycloidio/terraform v0.12.26-cy

go 1.14
