package redisenterprise

import "strings"

type AccessKeyType string

const (
	AccessKeyTypePrimary   AccessKeyType = "Primary"
	AccessKeyTypeSecondary AccessKeyType = "Secondary"
)

func PossibleValuesForAccessKeyType() []string {
	return []string{
		string(AccessKeyTypePrimary),
		string(AccessKeyTypeSecondary),
	}
}

func parseAccessKeyType(input string) (*AccessKeyType, error) {
	vals := map[string]AccessKeyType{
		"primary":   AccessKeyTypePrimary,
		"secondary": AccessKeyTypeSecondary,
	}
	if v, ok := vals[strings.ToLower(input)]; ok {
		return &v, nil
	}

	// otherwise presume it's an undefined value and best-effort it
	out := AccessKeyType(input)
	return &out, nil
}

type AofFrequency string

const (
	AofFrequencyAlways AofFrequency = "always"
	AofFrequencyOnes   AofFrequency = "1s"
)

func PossibleValuesForAofFrequency() []string {
	return []string{
		string(AofFrequencyAlways),
		string(AofFrequencyOnes),
	}
}

func parseAofFrequency(input string) (*AofFrequency, error) {
	vals := map[string]AofFrequency{
		"always": AofFrequencyAlways,
		"1s":     AofFrequencyOnes,
	}
	if v, ok := vals[strings.ToLower(input)]; ok {
		return &v, nil
	}

	// otherwise presume it's an undefined value and best-effort it
	out := AofFrequency(input)
	return &out, nil
}

type ClusteringPolicy string

const (
	ClusteringPolicyEnterpriseCluster ClusteringPolicy = "EnterpriseCluster"
	ClusteringPolicyOSSCluster        ClusteringPolicy = "OSSCluster"
)

func PossibleValuesForClusteringPolicy() []string {
	return []string{
		string(ClusteringPolicyEnterpriseCluster),
		string(ClusteringPolicyOSSCluster),
	}
}

func parseClusteringPolicy(input string) (*ClusteringPolicy, error) {
	vals := map[string]ClusteringPolicy{
		"enterprisecluster": ClusteringPolicyEnterpriseCluster,
		"osscluster":        ClusteringPolicyOSSCluster,
	}
	if v, ok := vals[strings.ToLower(input)]; ok {
		return &v, nil
	}

	// otherwise presume it's an undefined value and best-effort it
	out := ClusteringPolicy(input)
	return &out, nil
}

type EvictionPolicy string

const (
	EvictionPolicyAllKeysLFU     EvictionPolicy = "AllKeysLFU"
	EvictionPolicyAllKeysLRU     EvictionPolicy = "AllKeysLRU"
	EvictionPolicyAllKeysRandom  EvictionPolicy = "AllKeysRandom"
	EvictionPolicyNoEviction     EvictionPolicy = "NoEviction"
	EvictionPolicyVolatileLFU    EvictionPolicy = "VolatileLFU"
	EvictionPolicyVolatileLRU    EvictionPolicy = "VolatileLRU"
	EvictionPolicyVolatileRandom EvictionPolicy = "VolatileRandom"
	EvictionPolicyVolatileTTL    EvictionPolicy = "VolatileTTL"
)

func PossibleValuesForEvictionPolicy() []string {
	return []string{
		string(EvictionPolicyAllKeysLFU),
		string(EvictionPolicyAllKeysLRU),
		string(EvictionPolicyAllKeysRandom),
		string(EvictionPolicyNoEviction),
		string(EvictionPolicyVolatileLFU),
		string(EvictionPolicyVolatileLRU),
		string(EvictionPolicyVolatileRandom),
		string(EvictionPolicyVolatileTTL),
	}
}

func parseEvictionPolicy(input string) (*EvictionPolicy, error) {
	vals := map[string]EvictionPolicy{
		"allkeyslfu":     EvictionPolicyAllKeysLFU,
		"allkeyslru":     EvictionPolicyAllKeysLRU,
		"allkeysrandom":  EvictionPolicyAllKeysRandom,
		"noeviction":     EvictionPolicyNoEviction,
		"volatilelfu":    EvictionPolicyVolatileLFU,
		"volatilelru":    EvictionPolicyVolatileLRU,
		"volatilerandom": EvictionPolicyVolatileRandom,
		"volatilettl":    EvictionPolicyVolatileTTL,
	}
	if v, ok := vals[strings.ToLower(input)]; ok {
		return &v, nil
	}

	// otherwise presume it's an undefined value and best-effort it
	out := EvictionPolicy(input)
	return &out, nil
}

type LinkState string

const (
	LinkStateLinkFailed   LinkState = "LinkFailed"
	LinkStateLinked       LinkState = "Linked"
	LinkStateLinking      LinkState = "Linking"
	LinkStateUnlinkFailed LinkState = "UnlinkFailed"
	LinkStateUnlinking    LinkState = "Unlinking"
)

func PossibleValuesForLinkState() []string {
	return []string{
		string(LinkStateLinkFailed),
		string(LinkStateLinked),
		string(LinkStateLinking),
		string(LinkStateUnlinkFailed),
		string(LinkStateUnlinking),
	}
}

func parseLinkState(input string) (*LinkState, error) {
	vals := map[string]LinkState{
		"linkfailed":   LinkStateLinkFailed,
		"linked":       LinkStateLinked,
		"linking":      LinkStateLinking,
		"unlinkfailed": LinkStateUnlinkFailed,
		"unlinking":    LinkStateUnlinking,
	}
	if v, ok := vals[strings.ToLower(input)]; ok {
		return &v, nil
	}

	// otherwise presume it's an undefined value and best-effort it
	out := LinkState(input)
	return &out, nil
}

type PrivateEndpointConnectionProvisioningState string

const (
	PrivateEndpointConnectionProvisioningStateCreating  PrivateEndpointConnectionProvisioningState = "Creating"
	PrivateEndpointConnectionProvisioningStateDeleting  PrivateEndpointConnectionProvisioningState = "Deleting"
	PrivateEndpointConnectionProvisioningStateFailed    PrivateEndpointConnectionProvisioningState = "Failed"
	PrivateEndpointConnectionProvisioningStateSucceeded PrivateEndpointConnectionProvisioningState = "Succeeded"
)

func PossibleValuesForPrivateEndpointConnectionProvisioningState() []string {
	return []string{
		string(PrivateEndpointConnectionProvisioningStateCreating),
		string(PrivateEndpointConnectionProvisioningStateDeleting),
		string(PrivateEndpointConnectionProvisioningStateFailed),
		string(PrivateEndpointConnectionProvisioningStateSucceeded),
	}
}

func parsePrivateEndpointConnectionProvisioningState(input string) (*PrivateEndpointConnectionProvisioningState, error) {
	vals := map[string]PrivateEndpointConnectionProvisioningState{
		"creating":  PrivateEndpointConnectionProvisioningStateCreating,
		"deleting":  PrivateEndpointConnectionProvisioningStateDeleting,
		"failed":    PrivateEndpointConnectionProvisioningStateFailed,
		"succeeded": PrivateEndpointConnectionProvisioningStateSucceeded,
	}
	if v, ok := vals[strings.ToLower(input)]; ok {
		return &v, nil
	}

	// otherwise presume it's an undefined value and best-effort it
	out := PrivateEndpointConnectionProvisioningState(input)
	return &out, nil
}

type PrivateEndpointServiceConnectionStatus string

const (
	PrivateEndpointServiceConnectionStatusApproved PrivateEndpointServiceConnectionStatus = "Approved"
	PrivateEndpointServiceConnectionStatusPending  PrivateEndpointServiceConnectionStatus = "Pending"
	PrivateEndpointServiceConnectionStatusRejected PrivateEndpointServiceConnectionStatus = "Rejected"
)

func PossibleValuesForPrivateEndpointServiceConnectionStatus() []string {
	return []string{
		string(PrivateEndpointServiceConnectionStatusApproved),
		string(PrivateEndpointServiceConnectionStatusPending),
		string(PrivateEndpointServiceConnectionStatusRejected),
	}
}

func parsePrivateEndpointServiceConnectionStatus(input string) (*PrivateEndpointServiceConnectionStatus, error) {
	vals := map[string]PrivateEndpointServiceConnectionStatus{
		"approved": PrivateEndpointServiceConnectionStatusApproved,
		"pending":  PrivateEndpointServiceConnectionStatusPending,
		"rejected": PrivateEndpointServiceConnectionStatusRejected,
	}
	if v, ok := vals[strings.ToLower(input)]; ok {
		return &v, nil
	}

	// otherwise presume it's an undefined value and best-effort it
	out := PrivateEndpointServiceConnectionStatus(input)
	return &out, nil
}

type Protocol string

const (
	ProtocolEncrypted Protocol = "Encrypted"
	ProtocolPlaintext Protocol = "Plaintext"
)

func PossibleValuesForProtocol() []string {
	return []string{
		string(ProtocolEncrypted),
		string(ProtocolPlaintext),
	}
}

func parseProtocol(input string) (*Protocol, error) {
	vals := map[string]Protocol{
		"encrypted": ProtocolEncrypted,
		"plaintext": ProtocolPlaintext,
	}
	if v, ok := vals[strings.ToLower(input)]; ok {
		return &v, nil
	}

	// otherwise presume it's an undefined value and best-effort it
	out := Protocol(input)
	return &out, nil
}

type ProvisioningState string

const (
	ProvisioningStateCanceled  ProvisioningState = "Canceled"
	ProvisioningStateCreating  ProvisioningState = "Creating"
	ProvisioningStateDeleting  ProvisioningState = "Deleting"
	ProvisioningStateFailed    ProvisioningState = "Failed"
	ProvisioningStateSucceeded ProvisioningState = "Succeeded"
	ProvisioningStateUpdating  ProvisioningState = "Updating"
)

func PossibleValuesForProvisioningState() []string {
	return []string{
		string(ProvisioningStateCanceled),
		string(ProvisioningStateCreating),
		string(ProvisioningStateDeleting),
		string(ProvisioningStateFailed),
		string(ProvisioningStateSucceeded),
		string(ProvisioningStateUpdating),
	}
}

func parseProvisioningState(input string) (*ProvisioningState, error) {
	vals := map[string]ProvisioningState{
		"canceled":  ProvisioningStateCanceled,
		"creating":  ProvisioningStateCreating,
		"deleting":  ProvisioningStateDeleting,
		"failed":    ProvisioningStateFailed,
		"succeeded": ProvisioningStateSucceeded,
		"updating":  ProvisioningStateUpdating,
	}
	if v, ok := vals[strings.ToLower(input)]; ok {
		return &v, nil
	}

	// otherwise presume it's an undefined value and best-effort it
	out := ProvisioningState(input)
	return &out, nil
}

type RdbFrequency string

const (
	RdbFrequencyOneTwoh RdbFrequency = "12h"
	RdbFrequencyOneh    RdbFrequency = "1h"
	RdbFrequencySixh    RdbFrequency = "6h"
)

func PossibleValuesForRdbFrequency() []string {
	return []string{
		string(RdbFrequencyOneTwoh),
		string(RdbFrequencyOneh),
		string(RdbFrequencySixh),
	}
}

func parseRdbFrequency(input string) (*RdbFrequency, error) {
	vals := map[string]RdbFrequency{
		"12h": RdbFrequencyOneTwoh,
		"1h":  RdbFrequencyOneh,
		"6h":  RdbFrequencySixh,
	}
	if v, ok := vals[strings.ToLower(input)]; ok {
		return &v, nil
	}

	// otherwise presume it's an undefined value and best-effort it
	out := RdbFrequency(input)
	return &out, nil
}

type ResourceState string

const (
	ResourceStateCreateFailed  ResourceState = "CreateFailed"
	ResourceStateCreating      ResourceState = "Creating"
	ResourceStateDeleteFailed  ResourceState = "DeleteFailed"
	ResourceStateDeleting      ResourceState = "Deleting"
	ResourceStateDisableFailed ResourceState = "DisableFailed"
	ResourceStateDisabled      ResourceState = "Disabled"
	ResourceStateDisabling     ResourceState = "Disabling"
	ResourceStateEnableFailed  ResourceState = "EnableFailed"
	ResourceStateEnabling      ResourceState = "Enabling"
	ResourceStateRunning       ResourceState = "Running"
	ResourceStateUpdateFailed  ResourceState = "UpdateFailed"
	ResourceStateUpdating      ResourceState = "Updating"
)

func PossibleValuesForResourceState() []string {
	return []string{
		string(ResourceStateCreateFailed),
		string(ResourceStateCreating),
		string(ResourceStateDeleteFailed),
		string(ResourceStateDeleting),
		string(ResourceStateDisableFailed),
		string(ResourceStateDisabled),
		string(ResourceStateDisabling),
		string(ResourceStateEnableFailed),
		string(ResourceStateEnabling),
		string(ResourceStateRunning),
		string(ResourceStateUpdateFailed),
		string(ResourceStateUpdating),
	}
}

func parseResourceState(input string) (*ResourceState, error) {
	vals := map[string]ResourceState{
		"createfailed":  ResourceStateCreateFailed,
		"creating":      ResourceStateCreating,
		"deletefailed":  ResourceStateDeleteFailed,
		"deleting":      ResourceStateDeleting,
		"disablefailed": ResourceStateDisableFailed,
		"disabled":      ResourceStateDisabled,
		"disabling":     ResourceStateDisabling,
		"enablefailed":  ResourceStateEnableFailed,
		"enabling":      ResourceStateEnabling,
		"running":       ResourceStateRunning,
		"updatefailed":  ResourceStateUpdateFailed,
		"updating":      ResourceStateUpdating,
	}
	if v, ok := vals[strings.ToLower(input)]; ok {
		return &v, nil
	}

	// otherwise presume it's an undefined value and best-effort it
	out := ResourceState(input)
	return &out, nil
}

type SkuName string

const (
	SkuNameEnterpriseEFiveZero             SkuName = "Enterprise_E50"
	SkuNameEnterpriseEOneZero              SkuName = "Enterprise_E10"
	SkuNameEnterpriseEOneZeroZero          SkuName = "Enterprise_E100"
	SkuNameEnterpriseETwoZero              SkuName = "Enterprise_E20"
	SkuNameEnterpriseFlashFOneFiveZeroZero SkuName = "EnterpriseFlash_F1500"
	SkuNameEnterpriseFlashFSevenZeroZero   SkuName = "EnterpriseFlash_F700"
	SkuNameEnterpriseFlashFThreeZeroZero   SkuName = "EnterpriseFlash_F300"
)

func PossibleValuesForSkuName() []string {
	return []string{
		string(SkuNameEnterpriseEFiveZero),
		string(SkuNameEnterpriseEOneZero),
		string(SkuNameEnterpriseEOneZeroZero),
		string(SkuNameEnterpriseETwoZero),
		string(SkuNameEnterpriseFlashFOneFiveZeroZero),
		string(SkuNameEnterpriseFlashFSevenZeroZero),
		string(SkuNameEnterpriseFlashFThreeZeroZero),
	}
}

func parseSkuName(input string) (*SkuName, error) {
	vals := map[string]SkuName{
		"enterprise_e50":        SkuNameEnterpriseEFiveZero,
		"enterprise_e10":        SkuNameEnterpriseEOneZero,
		"enterprise_e100":       SkuNameEnterpriseEOneZeroZero,
		"enterprise_e20":        SkuNameEnterpriseETwoZero,
		"enterpriseflash_f1500": SkuNameEnterpriseFlashFOneFiveZeroZero,
		"enterpriseflash_f700":  SkuNameEnterpriseFlashFSevenZeroZero,
		"enterpriseflash_f300":  SkuNameEnterpriseFlashFThreeZeroZero,
	}
	if v, ok := vals[strings.ToLower(input)]; ok {
		return &v, nil
	}

	// otherwise presume it's an undefined value and best-effort it
	out := SkuName(input)
	return &out, nil
}

type TlsVersion string

const (
	TlsVersionOnePointOne  TlsVersion = "1.1"
	TlsVersionOnePointTwo  TlsVersion = "1.2"
	TlsVersionOnePointZero TlsVersion = "1.0"
)

func PossibleValuesForTlsVersion() []string {
	return []string{
		string(TlsVersionOnePointOne),
		string(TlsVersionOnePointTwo),
		string(TlsVersionOnePointZero),
	}
}

func parseTlsVersion(input string) (*TlsVersion, error) {
	vals := map[string]TlsVersion{
		"1.1": TlsVersionOnePointOne,
		"1.2": TlsVersionOnePointTwo,
		"1.0": TlsVersionOnePointZero,
	}
	if v, ok := vals[strings.ToLower(input)]; ok {
		return &v, nil
	}

	// otherwise presume it's an undefined value and best-effort it
	out := TlsVersion(input)
	return &out, nil
}
