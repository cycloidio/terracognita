package iscsitargets

import "strings"

type CreatedByType string

const (
	CreatedByTypeApplication     CreatedByType = "Application"
	CreatedByTypeKey             CreatedByType = "Key"
	CreatedByTypeManagedIdentity CreatedByType = "ManagedIdentity"
	CreatedByTypeUser            CreatedByType = "User"
)

func PossibleValuesForCreatedByType() []string {
	return []string{
		string(CreatedByTypeApplication),
		string(CreatedByTypeKey),
		string(CreatedByTypeManagedIdentity),
		string(CreatedByTypeUser),
	}
}

func parseCreatedByType(input string) (*CreatedByType, error) {
	vals := map[string]CreatedByType{
		"application":     CreatedByTypeApplication,
		"key":             CreatedByTypeKey,
		"managedidentity": CreatedByTypeManagedIdentity,
		"user":            CreatedByTypeUser,
	}
	if v, ok := vals[strings.ToLower(input)]; ok {
		return &v, nil
	}

	// otherwise presume it's an undefined value and best-effort it
	out := CreatedByType(input)
	return &out, nil
}

type IscsiTargetAclMode string

const (
	IscsiTargetAclModeDynamic IscsiTargetAclMode = "Dynamic"
	IscsiTargetAclModeStatic  IscsiTargetAclMode = "Static"
)

func PossibleValuesForIscsiTargetAclMode() []string {
	return []string{
		string(IscsiTargetAclModeDynamic),
		string(IscsiTargetAclModeStatic),
	}
}

func parseIscsiTargetAclMode(input string) (*IscsiTargetAclMode, error) {
	vals := map[string]IscsiTargetAclMode{
		"dynamic": IscsiTargetAclModeDynamic,
		"static":  IscsiTargetAclModeStatic,
	}
	if v, ok := vals[strings.ToLower(input)]; ok {
		return &v, nil
	}

	// otherwise presume it's an undefined value and best-effort it
	out := IscsiTargetAclMode(input)
	return &out, nil
}

type OperationalStatus string

const (
	OperationalStatusHealthy            OperationalStatus = "Healthy"
	OperationalStatusInvalid            OperationalStatus = "Invalid"
	OperationalStatusRunning            OperationalStatus = "Running"
	OperationalStatusStopped            OperationalStatus = "Stopped"
	OperationalStatusStoppedDeallocated OperationalStatus = "Stopped (deallocated)"
	OperationalStatusUnhealthy          OperationalStatus = "Unhealthy"
	OperationalStatusUnknown            OperationalStatus = "Unknown"
	OperationalStatusUpdating           OperationalStatus = "Updating"
)

func PossibleValuesForOperationalStatus() []string {
	return []string{
		string(OperationalStatusHealthy),
		string(OperationalStatusInvalid),
		string(OperationalStatusRunning),
		string(OperationalStatusStopped),
		string(OperationalStatusStoppedDeallocated),
		string(OperationalStatusUnhealthy),
		string(OperationalStatusUnknown),
		string(OperationalStatusUpdating),
	}
}

func parseOperationalStatus(input string) (*OperationalStatus, error) {
	vals := map[string]OperationalStatus{
		"healthy":               OperationalStatusHealthy,
		"invalid":               OperationalStatusInvalid,
		"running":               OperationalStatusRunning,
		"stopped":               OperationalStatusStopped,
		"stopped (deallocated)": OperationalStatusStoppedDeallocated,
		"unhealthy":             OperationalStatusUnhealthy,
		"unknown":               OperationalStatusUnknown,
		"updating":              OperationalStatusUpdating,
	}
	if v, ok := vals[strings.ToLower(input)]; ok {
		return &v, nil
	}

	// otherwise presume it's an undefined value and best-effort it
	out := OperationalStatus(input)
	return &out, nil
}

type ProvisioningStates string

const (
	ProvisioningStatesCanceled  ProvisioningStates = "Canceled"
	ProvisioningStatesCreating  ProvisioningStates = "Creating"
	ProvisioningStatesDeleting  ProvisioningStates = "Deleting"
	ProvisioningStatesFailed    ProvisioningStates = "Failed"
	ProvisioningStatesInvalid   ProvisioningStates = "Invalid"
	ProvisioningStatesPending   ProvisioningStates = "Pending"
	ProvisioningStatesSucceeded ProvisioningStates = "Succeeded"
	ProvisioningStatesUpdating  ProvisioningStates = "Updating"
)

func PossibleValuesForProvisioningStates() []string {
	return []string{
		string(ProvisioningStatesCanceled),
		string(ProvisioningStatesCreating),
		string(ProvisioningStatesDeleting),
		string(ProvisioningStatesFailed),
		string(ProvisioningStatesInvalid),
		string(ProvisioningStatesPending),
		string(ProvisioningStatesSucceeded),
		string(ProvisioningStatesUpdating),
	}
}

func parseProvisioningStates(input string) (*ProvisioningStates, error) {
	vals := map[string]ProvisioningStates{
		"canceled":  ProvisioningStatesCanceled,
		"creating":  ProvisioningStatesCreating,
		"deleting":  ProvisioningStatesDeleting,
		"failed":    ProvisioningStatesFailed,
		"invalid":   ProvisioningStatesInvalid,
		"pending":   ProvisioningStatesPending,
		"succeeded": ProvisioningStatesSucceeded,
		"updating":  ProvisioningStatesUpdating,
	}
	if v, ok := vals[strings.ToLower(input)]; ok {
		return &v, nil
	}

	// otherwise presume it's an undefined value and best-effort it
	out := ProvisioningStates(input)
	return &out, nil
}
