module github.com/cycloidio/terracognita

require (
	cloud.google.com/go/bigtable v1.0.0 // indirect
	cloud.google.com/go/storage v1.0.0 // indirect
	github.com/Azure/azure-sdk-for-go v33.2.0+incompatible
	github.com/Azure/go-autorest/autorest v0.9.3
	github.com/Azure/go-autorest/autorest/adal v0.8.1-0.20191028180845-3492b2aff503 // indirect
	github.com/aws/aws-sdk-go v1.25.4
	github.com/chr4/pwgen v1.1.0
	github.com/cycloidio/tfdocs v0.0.0-20200111145532-e6a80a93d7cc
	github.com/go-kit/kit v0.9.0
	github.com/golang/mock v1.3.1
	github.com/hashicorp/go-azure-helpers v0.9.0
	github.com/hashicorp/hcl v1.0.0
	github.com/hashicorp/hil v0.0.0-20190212132231-97b3a9cdfa93 // indirect
	github.com/hashicorp/terraform v0.12.8
	github.com/hashicorp/vault v1.0.3 // indirect
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.4.0
	github.com/terraform-providers/terraform-provider-aws v1.60.1-0.20191003145700-f8707a46c6ec
	github.com/terraform-providers/terraform-provider-azurerm v1.35.0
	github.com/terraform-providers/terraform-provider-google v1.20.1-0.20190924213132-8cb5c9efd9d7
	github.com/zclconf/go-cty v1.1.0
	golang.org/x/crypto v0.0.0-20200311171314-f7b00557c8c4 // indirect
	golang.org/x/exp v0.0.0-20190912063710-ac5d2bfcbfe0 // indirect
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/net v0.0.0-20200301022130-244492dfa37a // indirect
	golang.org/x/sys v0.0.0-20200302150141-5c8b2ff67527 // indirect
	golang.org/x/tools v0.0.0-20200312194400-c312e98713c2 // indirect
	google.golang.org/api v0.9.0
	k8s.io/apimachinery v0.0.0-20190213030929-f84a4639d8e8 // indirect
	k8s.io/klog v0.2.0 // indirect
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999

replace github.com/terraform-providers/terraform-provider-tls v2.1.0+incompatible => github.com/terraform-providers/terraform-provider-tls v1.2.1-0.20190816230231-0790c4b40281

go 1.11
