package checknameavailabilitydisasterrecoveryconfigs

import "strings"

type UnavailableReason string

const (
	UnavailableReasonInvalidName                           UnavailableReason = "InvalidName"
	UnavailableReasonNameInLockdown                        UnavailableReason = "NameInLockdown"
	UnavailableReasonNameInUse                             UnavailableReason = "NameInUse"
	UnavailableReasonNone                                  UnavailableReason = "None"
	UnavailableReasonSubscriptionIsDisabled                UnavailableReason = "SubscriptionIsDisabled"
	UnavailableReasonTooManyNamespaceInCurrentSubscription UnavailableReason = "TooManyNamespaceInCurrentSubscription"
)

func PossibleValuesForUnavailableReason() []string {
	return []string{
		string(UnavailableReasonInvalidName),
		string(UnavailableReasonNameInLockdown),
		string(UnavailableReasonNameInUse),
		string(UnavailableReasonNone),
		string(UnavailableReasonSubscriptionIsDisabled),
		string(UnavailableReasonTooManyNamespaceInCurrentSubscription),
	}
}

func parseUnavailableReason(input string) (*UnavailableReason, error) {
	vals := map[string]UnavailableReason{
		"invalidname":                           UnavailableReasonInvalidName,
		"nameinlockdown":                        UnavailableReasonNameInLockdown,
		"nameinuse":                             UnavailableReasonNameInUse,
		"none":                                  UnavailableReasonNone,
		"subscriptionisdisabled":                UnavailableReasonSubscriptionIsDisabled,
		"toomanynamespaceincurrentsubscription": UnavailableReasonTooManyNamespaceInCurrentSubscription,
	}
	if v, ok := vals[strings.ToLower(input)]; ok {
		return &v, nil
	}

	// otherwise presume it's an undefined value and best-effort it
	out := UnavailableReason(input)
	return &out, nil
}
