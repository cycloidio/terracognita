package tenants

import "strings"

type BillingType string

const (
	BillingTypeAuths              BillingType = "auths"
	BillingTypeMonthlyActiveUsers BillingType = "mau"
)

func PossibleValuesForBillingType() []string {
	return []string{
		string(BillingTypeAuths),
		string(BillingTypeMonthlyActiveUsers),
	}
}

func parseBillingType(input string) (*BillingType, error) {
	vals := map[string]BillingType{
		"auths": BillingTypeAuths,
		"mau":   BillingTypeMonthlyActiveUsers,
	}
	if v, ok := vals[strings.ToLower(input)]; ok {
		return &v, nil
	}

	// otherwise presume it's an undefined value and best-effort it
	out := BillingType(input)
	return &out, nil
}

type Location string

const (
	LocationAsiaPacific  Location = "Asia Pacific"
	LocationAustralia    Location = "Australia"
	LocationEurope       Location = "Europe"
	LocationGlobal       Location = "Global"
	LocationUnitedStates Location = "United States"
)

func PossibleValuesForLocation() []string {
	return []string{
		string(LocationAsiaPacific),
		string(LocationAustralia),
		string(LocationEurope),
		string(LocationGlobal),
		string(LocationUnitedStates),
	}
}

func parseLocation(input string) (*Location, error) {
	vals := map[string]Location{
		"asia pacific":  LocationAsiaPacific,
		"australia":     LocationAustralia,
		"europe":        LocationEurope,
		"global":        LocationGlobal,
		"united states": LocationUnitedStates,
	}
	if v, ok := vals[strings.ToLower(input)]; ok {
		return &v, nil
	}

	// otherwise presume it's an undefined value and best-effort it
	out := Location(input)
	return &out, nil
}

type SkuName string

const (
	SkuNamePremiumP1 SkuName = "PremiumP1"
	SkuNamePremiumP2 SkuName = "PremiumP2"
	SkuNameStandard  SkuName = "Standard"
)

func PossibleValuesForSkuName() []string {
	return []string{
		string(SkuNamePremiumP1),
		string(SkuNamePremiumP2),
		string(SkuNameStandard),
	}
}

func parseSkuName(input string) (*SkuName, error) {
	vals := map[string]SkuName{
		"premiump1": SkuNamePremiumP1,
		"premiump2": SkuNamePremiumP2,
		"standard":  SkuNameStandard,
	}
	if v, ok := vals[strings.ToLower(input)]; ok {
		return &v, nil
	}

	// otherwise presume it's an undefined value and best-effort it
	out := SkuName(input)
	return &out, nil
}

type SkuTier string

const (
	SkuTierA0 SkuTier = "A0"
)

func PossibleValuesForSkuTier() []string {
	return []string{
		string(SkuTierA0),
	}
}

func parseSkuTier(input string) (*SkuTier, error) {
	vals := map[string]SkuTier{
		"a0": SkuTierA0,
	}
	if v, ok := vals[strings.ToLower(input)]; ok {
		return &v, nil
	}

	// otherwise presume it's an undefined value and best-effort it
	out := SkuTier(input)
	return &out, nil
}
