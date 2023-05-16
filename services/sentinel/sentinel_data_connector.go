package sentinel

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/preview/securityinsight/mgmt/2021-09-01-preview/securityinsight"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/sentinel/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
)

func importSentinelDataConnector(expectKind securityinsight.DataConnectorKind) pluginsdk.ImporterFunc {
	return func(ctx context.Context, d *pluginsdk.ResourceData, meta interface{}) (data []*pluginsdk.ResourceData, err error) {
		id, err := parse.DataConnectorID(d.Id())
		if err != nil {
			return nil, err
		}

		client := meta.(*clients.Client).Sentinel.DataConnectorsClient
		resp, err := client.Get(ctx, id.ResourceGroup, id.WorkspaceName, id.Name)
		if err != nil {
			return nil, fmt.Errorf("retrieving Sentinel Alert Rule %q: %+v", id, err)
		}

		if err := assertDataConnectorKind(resp.Value, expectKind); err != nil {
			return nil, err
		}
		return []*pluginsdk.ResourceData{d}, nil
	}
}

func assertDataConnectorKind(dc securityinsight.BasicDataConnector, expectKind securityinsight.DataConnectorKind) error {
	var kind securityinsight.DataConnectorKind
	switch dc.(type) {
	case securityinsight.AADDataConnector:
		kind = securityinsight.DataConnectorKindAzureActiveDirectory
	case securityinsight.AATPDataConnector:
		kind = securityinsight.DataConnectorKindAzureAdvancedThreatProtection
	case securityinsight.ASCDataConnector:
		kind = securityinsight.DataConnectorKindAzureSecurityCenter
	case securityinsight.MCASDataConnector:
		kind = securityinsight.DataConnectorKindMicrosoftCloudAppSecurity
	case securityinsight.TIDataConnector:
		kind = securityinsight.DataConnectorKindThreatIntelligence
	case securityinsight.OfficeDataConnector:
		kind = securityinsight.DataConnectorKindOffice365
	case securityinsight.OfficeATPDataConnector:
		kind = securityinsight.DataConnectorKindOfficeATP
	case securityinsight.AwsCloudTrailDataConnector:
		kind = securityinsight.DataConnectorKindAmazonWebServicesCloudTrail
	case securityinsight.MDATPDataConnector:
		kind = securityinsight.DataConnectorKindMicrosoftDefenderAdvancedThreatProtection
	}
	if expectKind != kind {
		return fmt.Errorf("Sentinel Data Connector has mismatched kind, expected: %q, got %q", expectKind, kind)
	}
	return nil
}
