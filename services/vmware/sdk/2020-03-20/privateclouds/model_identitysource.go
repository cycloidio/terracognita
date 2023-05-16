package privateclouds

type IdentitySource struct {
	Alias           *string  `json:"alias,omitempty"`
	BaseGroupDN     *string  `json:"baseGroupDN,omitempty"`
	BaseUserDN      *string  `json:"baseUserDN,omitempty"`
	Domain          *string  `json:"domain,omitempty"`
	Name            *string  `json:"name,omitempty"`
	Password        *string  `json:"password,omitempty"`
	PrimaryServer   *string  `json:"primaryServer,omitempty"`
	SecondaryServer *string  `json:"secondaryServer,omitempty"`
	Ssl             *SslEnum `json:"ssl,omitempty"`
	Username        *string  `json:"username,omitempty"`
}
