module github.com/cycloidio/terracognita

go 1.16

require (
	github.com/Azure/azure-sdk-for-go v53.4.0+incompatible
	github.com/Azure/go-autorest/autorest v0.11.18
	github.com/adrg/xdg v0.2.3
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/aws/aws-sdk-go v1.38.37
	github.com/chr4/pwgen v1.1.0
	github.com/cycloidio/mxwriter v1.0.2
	github.com/cycloidio/tfdocs v0.0.0-20210713114615-6a7f069f11c3
	github.com/go-kit/kit v0.9.0
	github.com/golang/mock v1.4.4
	github.com/hashicorp/go-azure-helpers v0.15.0
	github.com/hashicorp/go-cty v1.4.1-0.20200414143053-d3edf31b6320
	github.com/hashicorp/hcl/v2 v2.8.2
	github.com/hashicorp/hil v0.0.0-20190212132231-97b3a9cdfa93 // indirect
	github.com/hashicorp/terraform v0.13.0
	github.com/hashicorp/terraform-plugin-go v0.3.0
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.6.1
	github.com/hashicorp/terraform-provider-google v1.20.1-0.20210510171431-a764cf3da527
	github.com/hashicorp/terraform-svchost v0.0.0-20200729002733-f050f53b9734 // indirect
	github.com/jinzhu/inflection v1.0.0
	github.com/keybase/go-crypto v0.0.0-20181127160227-255a5089e85a // indirect
	github.com/pascaldekloe/name v1.0.1
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.7.0
	github.com/terraform-providers/terraform-provider-aws v1.60.1-0.20210513231836-489654890359
	github.com/terraform-providers/terraform-provider-azurerm v1.44.1-0.20201029183808-d721bcc1bb55
	github.com/zclconf/go-cty v1.8.2
	golang.org/x/arch v0.0.0-20190312162104-788fe5ffcd8c // indirect
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616 // indirect
	golang.org/x/tools v0.1.4 // indirect
	google.golang.org/api v0.41.0
	google.golang.org/grpc v1.36.0
	gopkg.in/yaml.v2 v2.4.0
)

// Force an specific version if not the AWS provider does not compile
replace github.com/hashicorp/aws-sdk-go-base v0.6.0 => github.com/hashicorp/aws-sdk-go-base v0.5.0

// If we  go to the 1.5.0 then github.com/hashicorp/terraform-plugin-test/ will break
// as go-getter introduced a break from 1.4 -> 1.5
replace github.com/hashicorp/go-getter v1.5.0 => github.com/hashicorp/go-getter v1.4.0

// To remove the panic issue of using TF
replace github.com/hashicorp/terraform => github.com/cycloidio/terraform v0.13.5-cy

// Fork of Azurerm that has the V2 of the SDK
replace github.com/terraform-providers/terraform-provider-azurerm => github.com/cycloidio/terraform-provider-azurerm v1.44.1-0.20210517111036-df0beb5af9c3
