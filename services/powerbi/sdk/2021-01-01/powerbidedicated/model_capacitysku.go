package powerbidedicated

type CapacitySku struct {
	Name string           `json:"name"`
	Tier *CapacitySkuTier `json:"tier,omitempty"`
}
