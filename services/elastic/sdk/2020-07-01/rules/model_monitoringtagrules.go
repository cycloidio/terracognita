package rules

type MonitoringTagRules struct {
	Id         *string                       `json:"id,omitempty"`
	Name       *string                       `json:"name,omitempty"`
	Properties *MonitoringTagRulesProperties `json:"properties,omitempty"`
	SystemData *SystemData                   `json:"systemData,omitempty"`
	Type       *string                       `json:"type,omitempty"`
}
