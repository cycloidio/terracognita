module github.com/cycloidio/terracognita

require (
	github.com/aws/aws-sdk-go v1.25.4
	github.com/chr4/pwgen v1.1.0
	github.com/cycloidio/raws v1.0.1
	github.com/go-kit/kit v0.9.0
	github.com/golang/mock v1.3.1
	github.com/hashicorp/hcl v1.0.0
	github.com/hashicorp/hil v0.0.0-20190212132231-97b3a9cdfa93 // indirect
	github.com/hashicorp/terraform v0.12.7
	github.com/hashicorp/vault v1.0.3 // indirect
	github.com/keybase/go-crypto v0.0.0-20181127160227-255a5089e85a // indirect
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.4.0
	github.com/terraform-providers/terraform-provider-aws v1.60.1-0.20191003145700-f8707a46c6ec
	k8s.io/apimachinery v0.0.0-20190213030929-f84a4639d8e8 // indirect
	k8s.io/klog v0.2.0 // indirect
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999

replace github.com/terraform-providers/terraform-provider-tls v2.1.0+incompatible => github.com/terraform-providers/terraform-provider-tls v1.2.1-0.20190816230231-0790c4b40281

go 1.11
