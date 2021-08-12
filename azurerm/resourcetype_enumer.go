// Code generated by "enumer -type ResourceType -addprefix azurerm_ -transform snake -linecomment"; DO NOT EDIT.

package azurerm

import (
	"fmt"
)

const _ResourceTypeName = "azurerm_resource_groupazurerm_virtual_machineazurerm_virtual_machine_extensionazurerm_virtual_machine_scale_setazurerm_virtual_networkazurerm_availability_setazurerm_imageazurerm_subnetazurerm_network_interfaceazurerm_network_security_groupazurerm_application_gatewayazurerm_application_security_groupazurerm_ddos_protection_planazurerm_azure_firewallazurerm_local_network_gatewayazurerm_nat_gatewayazurerm_profileazurerm_security_ruleazurerm_public_ip_addressazurerm_public_ip_prefixazurerm_routeazurerm_route_tableazurerm_virtual_network_gatewayazurerm_virtual_network_gateway_connectionazurerm_virtual_network_peeringazurerm_web_application_firewall_policyazurerm_virtual_desktop_host_poolazurerm_virtual_desktop_application_groupazurerm_logic_app_workflowazurerm_logic_app_trigger_customazurerm_logic_app_action_customazurerm_container_registryazurerm_container_registry_webhookazurerm_storage_accountazurerm_storage_queueazurerm_storage_file_shareazurerm_storage_tableazurerm_mariadb_configurationazurerm_mariadb_databaseazurerm_mariadb_firewall_ruleazurerm_mariadb_serverazurerm_mariadb_virtual_network_ruleazurerm_mysql_configurationazurerm_mysql_databaseazurerm_mysql_firewall_ruleazurerm_mysql_serverazurerm_mysql_virtual_network_ruleazurerm_postgresql_configurationazurerm_postgresql_databaseazurerm_postgresql_firewall_ruleazurerm_postgresql_serverazurerm_postgresql_virtual_network_ruleazurerm_sql_elastic_poolazurerm_sql_databaseazurerm_sql_firewall_ruleazurerm_sql_server"

var _ResourceTypeIndex = [...]uint16{0, 22, 45, 78, 111, 134, 158, 171, 185, 210, 240, 267, 301, 329, 351, 380, 399, 414, 435, 460, 484, 497, 516, 547, 589, 620, 659, 692, 733, 759, 791, 822, 848, 882, 905, 926, 952, 973, 1002, 1026, 1055, 1077, 1113, 1140, 1162, 1189, 1209, 1243, 1275, 1302, 1334, 1359, 1398, 1422, 1442, 1467, 1485}

const _ResourceTypeLowerName = "azurerm_resource_groupazurerm_virtual_machineazurerm_virtual_machine_extensionazurerm_virtual_machine_scale_setazurerm_virtual_networkazurerm_availability_setazurerm_imageazurerm_subnetazurerm_network_interfaceazurerm_network_security_groupazurerm_application_gatewayazurerm_application_security_groupazurerm_ddos_protection_planazurerm_azure_firewallazurerm_local_network_gatewayazurerm_nat_gatewayazurerm_profileazurerm_security_ruleazurerm_public_ip_addressazurerm_public_ip_prefixazurerm_routeazurerm_route_tableazurerm_virtual_network_gatewayazurerm_virtual_network_gateway_connectionazurerm_virtual_network_peeringazurerm_web_application_firewall_policyazurerm_virtual_desktop_host_poolazurerm_virtual_desktop_application_groupazurerm_logic_app_workflowazurerm_logic_app_trigger_customazurerm_logic_app_action_customazurerm_container_registryazurerm_container_registry_webhookazurerm_storage_accountazurerm_storage_queueazurerm_storage_file_shareazurerm_storage_tableazurerm_mariadb_configurationazurerm_mariadb_databaseazurerm_mariadb_firewall_ruleazurerm_mariadb_serverazurerm_mariadb_virtual_network_ruleazurerm_mysql_configurationazurerm_mysql_databaseazurerm_mysql_firewall_ruleazurerm_mysql_serverazurerm_mysql_virtual_network_ruleazurerm_postgresql_configurationazurerm_postgresql_databaseazurerm_postgresql_firewall_ruleazurerm_postgresql_serverazurerm_postgresql_virtual_network_ruleazurerm_sql_elastic_poolazurerm_sql_databaseazurerm_sql_firewall_ruleazurerm_sql_server"

func (i ResourceType) String() string {
	if i < 0 || i >= ResourceType(len(_ResourceTypeIndex)-1) {
		return fmt.Sprintf("ResourceType(%d)", i)
	}
	return _ResourceTypeName[_ResourceTypeIndex[i]:_ResourceTypeIndex[i+1]]
}

// An "invalid array index" compiler error signifies that the constant values have changed.
// Re-run the stringer command to generate them again.
func _ResourceTypeNoOp() {
	var x [1]struct{}
	_ = x[ResourceGroup-(0)]
	_ = x[VirtualMachine-(1)]
	_ = x[VirtualMachineExtension-(2)]
	_ = x[VirtualMachineScaleSet-(3)]
	_ = x[VirtualNetwork-(4)]
	_ = x[AvailabilitySet-(5)]
	_ = x[Image-(6)]
	_ = x[Subnet-(7)]
	_ = x[NetworkInterface-(8)]
	_ = x[NetworkSecurityGroup-(9)]
	_ = x[ApplicationGateway-(10)]
	_ = x[ApplicationSecurityGroup-(11)]
	_ = x[DdosProtectionPlan-(12)]
	_ = x[AzureFirewall-(13)]
	_ = x[LocalNetworkGateway-(14)]
	_ = x[NatGateway-(15)]
	_ = x[Profile-(16)]
	_ = x[SecurityRule-(17)]
	_ = x[PublicIPAddress-(18)]
	_ = x[PublicIPPrefix-(19)]
	_ = x[Route-(20)]
	_ = x[RouteTable-(21)]
	_ = x[VirtualNetworkGateway-(22)]
	_ = x[VirtualNetworkGatewayConnection-(23)]
	_ = x[VirtualNetworkPeering-(24)]
	_ = x[WebApplicationFirewallPolicy-(25)]
	_ = x[VirtualDesktopHostPool-(26)]
	_ = x[VirtualDesktopApplicationGroup-(27)]
	_ = x[LogicAppWorkflow-(28)]
	_ = x[LogicAppTriggerCustom-(29)]
	_ = x[LogicAppActionCustom-(30)]
	_ = x[ContainerRegistry-(31)]
	_ = x[ContainerRegistryWebhook-(32)]
	_ = x[StorageAccount-(33)]
	_ = x[StorageQueue-(34)]
	_ = x[StorageFileShare-(35)]
	_ = x[StorageTable-(36)]
	_ = x[MariadbConfiguration-(37)]
	_ = x[MariadbDatabase-(38)]
	_ = x[MariadbFirewallRule-(39)]
	_ = x[MariadbServer-(40)]
	_ = x[MariadbVirtualNetworkRule-(41)]
	_ = x[MysqlConfiguration-(42)]
	_ = x[MysqlDatabase-(43)]
	_ = x[MysqlFirewallRule-(44)]
	_ = x[MysqlServer-(45)]
	_ = x[MysqlVirtualNetworkRule-(46)]
	_ = x[PostgresqlConfiguration-(47)]
	_ = x[PostgresqlDatabase-(48)]
	_ = x[PostgresqlFirewallRule-(49)]
	_ = x[PostgresqlServer-(50)]
	_ = x[PostgresqlVirtualNetworkRule-(51)]
	_ = x[SQLElasticPool-(52)]
	_ = x[SQLDatabase-(53)]
	_ = x[SQLFirewallRule-(54)]
	_ = x[SQLServer-(55)]
}

var _ResourceTypeValues = []ResourceType{ResourceGroup, VirtualMachine, VirtualMachineExtension, VirtualMachineScaleSet, VirtualNetwork, AvailabilitySet, Image, Subnet, NetworkInterface, NetworkSecurityGroup, ApplicationGateway, ApplicationSecurityGroup, DdosProtectionPlan, AzureFirewall, LocalNetworkGateway, NatGateway, Profile, SecurityRule, PublicIPAddress, PublicIPPrefix, Route, RouteTable, VirtualNetworkGateway, VirtualNetworkGatewayConnection, VirtualNetworkPeering, WebApplicationFirewallPolicy, VirtualDesktopHostPool, VirtualDesktopApplicationGroup, LogicAppWorkflow, LogicAppTriggerCustom, LogicAppActionCustom, ContainerRegistry, ContainerRegistryWebhook, StorageAccount, StorageQueue, StorageFileShare, StorageTable, MariadbConfiguration, MariadbDatabase, MariadbFirewallRule, MariadbServer, MariadbVirtualNetworkRule, MysqlConfiguration, MysqlDatabase, MysqlFirewallRule, MysqlServer, MysqlVirtualNetworkRule, PostgresqlConfiguration, PostgresqlDatabase, PostgresqlFirewallRule, PostgresqlServer, PostgresqlVirtualNetworkRule, SQLElasticPool, SQLDatabase, SQLFirewallRule, SQLServer}

var _ResourceTypeNameToValueMap = map[string]ResourceType{
	_ResourceTypeName[0:22]:           ResourceGroup,
	_ResourceTypeLowerName[0:22]:      ResourceGroup,
	_ResourceTypeName[22:45]:          VirtualMachine,
	_ResourceTypeLowerName[22:45]:     VirtualMachine,
	_ResourceTypeName[45:78]:          VirtualMachineExtension,
	_ResourceTypeLowerName[45:78]:     VirtualMachineExtension,
	_ResourceTypeName[78:111]:         VirtualMachineScaleSet,
	_ResourceTypeLowerName[78:111]:    VirtualMachineScaleSet,
	_ResourceTypeName[111:134]:        VirtualNetwork,
	_ResourceTypeLowerName[111:134]:   VirtualNetwork,
	_ResourceTypeName[134:158]:        AvailabilitySet,
	_ResourceTypeLowerName[134:158]:   AvailabilitySet,
	_ResourceTypeName[158:171]:        Image,
	_ResourceTypeLowerName[158:171]:   Image,
	_ResourceTypeName[171:185]:        Subnet,
	_ResourceTypeLowerName[171:185]:   Subnet,
	_ResourceTypeName[185:210]:        NetworkInterface,
	_ResourceTypeLowerName[185:210]:   NetworkInterface,
	_ResourceTypeName[210:240]:        NetworkSecurityGroup,
	_ResourceTypeLowerName[210:240]:   NetworkSecurityGroup,
	_ResourceTypeName[240:267]:        ApplicationGateway,
	_ResourceTypeLowerName[240:267]:   ApplicationGateway,
	_ResourceTypeName[267:301]:        ApplicationSecurityGroup,
	_ResourceTypeLowerName[267:301]:   ApplicationSecurityGroup,
	_ResourceTypeName[301:329]:        DdosProtectionPlan,
	_ResourceTypeLowerName[301:329]:   DdosProtectionPlan,
	_ResourceTypeName[329:351]:        AzureFirewall,
	_ResourceTypeLowerName[329:351]:   AzureFirewall,
	_ResourceTypeName[351:380]:        LocalNetworkGateway,
	_ResourceTypeLowerName[351:380]:   LocalNetworkGateway,
	_ResourceTypeName[380:399]:        NatGateway,
	_ResourceTypeLowerName[380:399]:   NatGateway,
	_ResourceTypeName[399:414]:        Profile,
	_ResourceTypeLowerName[399:414]:   Profile,
	_ResourceTypeName[414:435]:        SecurityRule,
	_ResourceTypeLowerName[414:435]:   SecurityRule,
	_ResourceTypeName[435:460]:        PublicIPAddress,
	_ResourceTypeLowerName[435:460]:   PublicIPAddress,
	_ResourceTypeName[460:484]:        PublicIPPrefix,
	_ResourceTypeLowerName[460:484]:   PublicIPPrefix,
	_ResourceTypeName[484:497]:        Route,
	_ResourceTypeLowerName[484:497]:   Route,
	_ResourceTypeName[497:516]:        RouteTable,
	_ResourceTypeLowerName[497:516]:   RouteTable,
	_ResourceTypeName[516:547]:        VirtualNetworkGateway,
	_ResourceTypeLowerName[516:547]:   VirtualNetworkGateway,
	_ResourceTypeName[547:589]:        VirtualNetworkGatewayConnection,
	_ResourceTypeLowerName[547:589]:   VirtualNetworkGatewayConnection,
	_ResourceTypeName[589:620]:        VirtualNetworkPeering,
	_ResourceTypeLowerName[589:620]:   VirtualNetworkPeering,
	_ResourceTypeName[620:659]:        WebApplicationFirewallPolicy,
	_ResourceTypeLowerName[620:659]:   WebApplicationFirewallPolicy,
	_ResourceTypeName[659:692]:        VirtualDesktopHostPool,
	_ResourceTypeLowerName[659:692]:   VirtualDesktopHostPool,
	_ResourceTypeName[692:733]:        VirtualDesktopApplicationGroup,
	_ResourceTypeLowerName[692:733]:   VirtualDesktopApplicationGroup,
	_ResourceTypeName[733:759]:        LogicAppWorkflow,
	_ResourceTypeLowerName[733:759]:   LogicAppWorkflow,
	_ResourceTypeName[759:791]:        LogicAppTriggerCustom,
	_ResourceTypeLowerName[759:791]:   LogicAppTriggerCustom,
	_ResourceTypeName[791:822]:        LogicAppActionCustom,
	_ResourceTypeLowerName[791:822]:   LogicAppActionCustom,
	_ResourceTypeName[822:848]:        ContainerRegistry,
	_ResourceTypeLowerName[822:848]:   ContainerRegistry,
	_ResourceTypeName[848:882]:        ContainerRegistryWebhook,
	_ResourceTypeLowerName[848:882]:   ContainerRegistryWebhook,
	_ResourceTypeName[882:905]:        StorageAccount,
	_ResourceTypeLowerName[882:905]:   StorageAccount,
	_ResourceTypeName[905:926]:        StorageQueue,
	_ResourceTypeLowerName[905:926]:   StorageQueue,
	_ResourceTypeName[926:952]:        StorageFileShare,
	_ResourceTypeLowerName[926:952]:   StorageFileShare,
	_ResourceTypeName[952:973]:        StorageTable,
	_ResourceTypeLowerName[952:973]:   StorageTable,
	_ResourceTypeName[973:1002]:       MariadbConfiguration,
	_ResourceTypeLowerName[973:1002]:  MariadbConfiguration,
	_ResourceTypeName[1002:1026]:      MariadbDatabase,
	_ResourceTypeLowerName[1002:1026]: MariadbDatabase,
	_ResourceTypeName[1026:1055]:      MariadbFirewallRule,
	_ResourceTypeLowerName[1026:1055]: MariadbFirewallRule,
	_ResourceTypeName[1055:1077]:      MariadbServer,
	_ResourceTypeLowerName[1055:1077]: MariadbServer,
	_ResourceTypeName[1077:1113]:      MariadbVirtualNetworkRule,
	_ResourceTypeLowerName[1077:1113]: MariadbVirtualNetworkRule,
	_ResourceTypeName[1113:1140]:      MysqlConfiguration,
	_ResourceTypeLowerName[1113:1140]: MysqlConfiguration,
	_ResourceTypeName[1140:1162]:      MysqlDatabase,
	_ResourceTypeLowerName[1140:1162]: MysqlDatabase,
	_ResourceTypeName[1162:1189]:      MysqlFirewallRule,
	_ResourceTypeLowerName[1162:1189]: MysqlFirewallRule,
	_ResourceTypeName[1189:1209]:      MysqlServer,
	_ResourceTypeLowerName[1189:1209]: MysqlServer,
	_ResourceTypeName[1209:1243]:      MysqlVirtualNetworkRule,
	_ResourceTypeLowerName[1209:1243]: MysqlVirtualNetworkRule,
	_ResourceTypeName[1243:1275]:      PostgresqlConfiguration,
	_ResourceTypeLowerName[1243:1275]: PostgresqlConfiguration,
	_ResourceTypeName[1275:1302]:      PostgresqlDatabase,
	_ResourceTypeLowerName[1275:1302]: PostgresqlDatabase,
	_ResourceTypeName[1302:1334]:      PostgresqlFirewallRule,
	_ResourceTypeLowerName[1302:1334]: PostgresqlFirewallRule,
	_ResourceTypeName[1334:1359]:      PostgresqlServer,
	_ResourceTypeLowerName[1334:1359]: PostgresqlServer,
	_ResourceTypeName[1359:1398]:      PostgresqlVirtualNetworkRule,
	_ResourceTypeLowerName[1359:1398]: PostgresqlVirtualNetworkRule,
	_ResourceTypeName[1398:1422]:      SQLElasticPool,
	_ResourceTypeLowerName[1398:1422]: SQLElasticPool,
	_ResourceTypeName[1422:1442]:      SQLDatabase,
	_ResourceTypeLowerName[1422:1442]: SQLDatabase,
	_ResourceTypeName[1442:1467]:      SQLFirewallRule,
	_ResourceTypeLowerName[1442:1467]: SQLFirewallRule,
	_ResourceTypeName[1467:1485]:      SQLServer,
	_ResourceTypeLowerName[1467:1485]: SQLServer,
}

var _ResourceTypeNames = []string{
	_ResourceTypeName[0:22],
	_ResourceTypeName[22:45],
	_ResourceTypeName[45:78],
	_ResourceTypeName[78:111],
	_ResourceTypeName[111:134],
	_ResourceTypeName[134:158],
	_ResourceTypeName[158:171],
	_ResourceTypeName[171:185],
	_ResourceTypeName[185:210],
	_ResourceTypeName[210:240],
	_ResourceTypeName[240:267],
	_ResourceTypeName[267:301],
	_ResourceTypeName[301:329],
	_ResourceTypeName[329:351],
	_ResourceTypeName[351:380],
	_ResourceTypeName[380:399],
	_ResourceTypeName[399:414],
	_ResourceTypeName[414:435],
	_ResourceTypeName[435:460],
	_ResourceTypeName[460:484],
	_ResourceTypeName[484:497],
	_ResourceTypeName[497:516],
	_ResourceTypeName[516:547],
	_ResourceTypeName[547:589],
	_ResourceTypeName[589:620],
	_ResourceTypeName[620:659],
	_ResourceTypeName[659:692],
	_ResourceTypeName[692:733],
	_ResourceTypeName[733:759],
	_ResourceTypeName[759:791],
	_ResourceTypeName[791:822],
	_ResourceTypeName[822:848],
	_ResourceTypeName[848:882],
	_ResourceTypeName[882:905],
	_ResourceTypeName[905:926],
	_ResourceTypeName[926:952],
	_ResourceTypeName[952:973],
	_ResourceTypeName[973:1002],
	_ResourceTypeName[1002:1026],
	_ResourceTypeName[1026:1055],
	_ResourceTypeName[1055:1077],
	_ResourceTypeName[1077:1113],
	_ResourceTypeName[1113:1140],
	_ResourceTypeName[1140:1162],
	_ResourceTypeName[1162:1189],
	_ResourceTypeName[1189:1209],
	_ResourceTypeName[1209:1243],
	_ResourceTypeName[1243:1275],
	_ResourceTypeName[1275:1302],
	_ResourceTypeName[1302:1334],
	_ResourceTypeName[1334:1359],
	_ResourceTypeName[1359:1398],
	_ResourceTypeName[1398:1422],
	_ResourceTypeName[1422:1442],
	_ResourceTypeName[1442:1467],
	_ResourceTypeName[1467:1485],
}

// ResourceTypeString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func ResourceTypeString(s string) (ResourceType, error) {
	if val, ok := _ResourceTypeNameToValueMap[s]; ok {
		return val, nil
	}

	return 0, fmt.Errorf("%s does not belong to ResourceType values", s)
}

// ResourceTypeValues returns all values of the enum
func ResourceTypeValues() []ResourceType {
	return _ResourceTypeValues
}

// ResourceTypeStrings returns a slice of all String values of the enum
func ResourceTypeStrings() []string {
	strs := make([]string, len(_ResourceTypeNames))
	copy(strs, _ResourceTypeNames)
	return strs
}

// IsAResourceType returns "true" if the value is listed in the enum definition. "false" otherwise
func (i ResourceType) IsAResourceType() bool {
	for _, v := range _ResourceTypeValues {
		if i == v {
			return true
		}
	}
	return false
}
