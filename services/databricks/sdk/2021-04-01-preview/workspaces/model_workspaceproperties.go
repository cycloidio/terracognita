package workspaces

import (
	"time"

	"github.com/hashicorp/go-azure-helpers/lang/dates"
)

type WorkspaceProperties struct {
	Authorizations             *[]WorkspaceProviderAuthorization `json:"authorizations,omitempty"`
	CreatedBy                  *CreatedBy                        `json:"createdBy,omitempty"`
	CreatedDateTime            *string                           `json:"createdDateTime,omitempty"`
	Encryption                 *WorkspacePropertiesEncryption    `json:"encryption,omitempty"`
	ManagedResourceGroupId     string                            `json:"managedResourceGroupId"`
	Parameters                 *WorkspaceCustomParameters        `json:"parameters,omitempty"`
	PrivateEndpointConnections *[]PrivateEndpointConnection      `json:"privateEndpointConnections,omitempty"`
	ProvisioningState          *ProvisioningState                `json:"provisioningState,omitempty"`
	PublicNetworkAccess        *PublicNetworkAccess              `json:"publicNetworkAccess,omitempty"`
	RequiredNsgRules           *RequiredNsgRules                 `json:"requiredNsgRules,omitempty"`
	StorageAccountIdentity     *ManagedIdentityConfiguration     `json:"storageAccountIdentity,omitempty"`
	UiDefinitionUri            *string                           `json:"uiDefinitionUri,omitempty"`
	UpdatedBy                  *CreatedBy                        `json:"updatedBy,omitempty"`
	WorkspaceId                *string                           `json:"workspaceId,omitempty"`
	WorkspaceUrl               *string                           `json:"workspaceUrl,omitempty"`
}

func (o WorkspaceProperties) GetCreatedDateTimeAsTime() (*time.Time, error) {
	if o.CreatedDateTime == nil {
		return nil, nil
	}
	return dates.ParseAsFormat(o.CreatedDateTime, "2006-01-02T15:04:05Z07:00")
}

func (o WorkspaceProperties) SetCreatedDateTimeAsTime(input time.Time) {
	formatted := input.Format("2006-01-02T15:04:05Z07:00")
	o.CreatedDateTime = &formatted
}
