// Code generated by "enumer -type ResourceType -addprefix azurerm_ -transform snake -linecomment"; DO NOT EDIT.

package azurerm

import (
	"fmt"
	"strings"
)

const _ResourceTypeName = "azurerm_resource_groupazurerm_virtual_machineazurerm_virtual_machine_extensionazurerm_virtual_machine_scale_setazurerm_virtual_networkazurerm_availability_setazurerm_imageazurerm_subnetazurerm_network_interfaceazurerm_network_security_groupazurerm_application_gatewayazurerm_application_security_groupazurerm_network_ddos_protection_planazurerm_firewallazurerm_local_network_gatewayazurerm_nat_gatewayazurerm_network_profileazurerm_network_security_ruleazurerm_public_ipazurerm_public_ip_prefixazurerm_routeazurerm_route_tableazurerm_virtual_network_gatewayazurerm_virtual_network_gateway_connectionazurerm_virtual_network_peeringazurerm_web_application_firewall_policyazurerm_virtual_desktop_host_poolazurerm_virtual_desktop_application_groupazurerm_logic_app_workflowazurerm_logic_app_trigger_customazurerm_logic_app_action_customazurerm_container_registryazurerm_container_registry_webhookazurerm_storage_accountazurerm_storage_queueazurerm_storage_shareazurerm_storage_tableazurerm_storage_blobazurerm_mariadb_configurationazurerm_mariadb_databaseazurerm_mariadb_firewall_ruleazurerm_mariadb_serverazurerm_mariadb_virtual_network_ruleazurerm_mysql_configurationazurerm_mysql_databaseazurerm_mysql_firewall_ruleazurerm_mysql_serverazurerm_mysql_virtual_network_ruleazurerm_postgresql_configurationazurerm_postgresql_databaseazurerm_postgresql_firewall_ruleazurerm_postgresql_serverazurerm_postgresql_virtual_network_ruleazurerm_sql_elasticpoolazurerm_sql_databaseazurerm_sql_firewall_ruleazurerm_sql_server"

var _ResourceTypeIndex = [...]uint16{0, 22, 45, 78, 111, 134, 158, 171, 185, 210, 240, 267, 301, 337, 353, 382, 401, 424, 453, 470, 494, 507, 526, 557, 599, 630, 669, 702, 743, 769, 801, 832, 858, 892, 915, 936, 957, 978, 998, 1027, 1051, 1080, 1102, 1138, 1165, 1187, 1214, 1234, 1268, 1300, 1327, 1359, 1384, 1423, 1446, 1466, 1491, 1509}

const _ResourceTypeLowerName = "azurerm_resource_groupazurerm_virtual_machineazurerm_virtual_machine_extensionazurerm_virtual_machine_scale_setazurerm_virtual_networkazurerm_availability_setazurerm_imageazurerm_subnetazurerm_network_interfaceazurerm_network_security_groupazurerm_application_gatewayazurerm_application_security_groupazurerm_network_ddos_protection_planazurerm_firewallazurerm_local_network_gatewayazurerm_nat_gatewayazurerm_network_profileazurerm_network_security_ruleazurerm_public_ipazurerm_public_ip_prefixazurerm_routeazurerm_route_tableazurerm_virtual_network_gatewayazurerm_virtual_network_gateway_connectionazurerm_virtual_network_peeringazurerm_web_application_firewall_policyazurerm_virtual_desktop_host_poolazurerm_virtual_desktop_application_groupazurerm_logic_app_workflowazurerm_logic_app_trigger_customazurerm_logic_app_action_customazurerm_container_registryazurerm_container_registry_webhookazurerm_storage_accountazurerm_storage_queueazurerm_storage_shareazurerm_storage_tableazurerm_storage_blobazurerm_mariadb_configurationazurerm_mariadb_databaseazurerm_mariadb_firewall_ruleazurerm_mariadb_serverazurerm_mariadb_virtual_network_ruleazurerm_mysql_configurationazurerm_mysql_databaseazurerm_mysql_firewall_ruleazurerm_mysql_serverazurerm_mysql_virtual_network_ruleazurerm_postgresql_configurationazurerm_postgresql_databaseazurerm_postgresql_firewall_ruleazurerm_postgresql_serverazurerm_postgresql_virtual_network_ruleazurerm_sql_elasticpoolazurerm_sql_databaseazurerm_sql_firewall_ruleazurerm_sql_server"

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
	_ = x[NetworkDdosProtectionPlan-(12)]
	_ = x[Firewall-(13)]
	_ = x[LocalNetworkGateway-(14)]
	_ = x[NatGateway-(15)]
	_ = x[NetworkProfile-(16)]
	_ = x[NetworkSecurityRule-(17)]
	_ = x[PublicIP-(18)]
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
	_ = x[StorageShare-(35)]
	_ = x[StorageTable-(36)]
	_ = x[StorageBlob-(37)]
	_ = x[MariadbConfiguration-(38)]
	_ = x[MariadbDatabase-(39)]
	_ = x[MariadbFirewallRule-(40)]
	_ = x[MariadbServer-(41)]
	_ = x[MariadbVirtualNetworkRule-(42)]
	_ = x[MysqlConfiguration-(43)]
	_ = x[MysqlDatabase-(44)]
	_ = x[MysqlFirewallRule-(45)]
	_ = x[MysqlServer-(46)]
	_ = x[MysqlVirtualNetworkRule-(47)]
	_ = x[PostgresqlConfiguration-(48)]
	_ = x[PostgresqlDatabase-(49)]
	_ = x[PostgresqlFirewallRule-(50)]
	_ = x[PostgresqlServer-(51)]
	_ = x[PostgresqlVirtualNetworkRule-(52)]
	_ = x[SQLElasticPool-(53)]
	_ = x[SQLDatabase-(54)]
	_ = x[SQLFirewallRule-(55)]
	_ = x[SQLServer-(56)]
}

var _ResourceTypeValues = []ResourceType{ResourceGroup, VirtualMachine, VirtualMachineExtension, VirtualMachineScaleSet, VirtualNetwork, AvailabilitySet, Image, Subnet, NetworkInterface, NetworkSecurityGroup, ApplicationGateway, ApplicationSecurityGroup, NetworkDdosProtectionPlan, Firewall, LocalNetworkGateway, NatGateway, NetworkProfile, NetworkSecurityRule, PublicIP, PublicIPPrefix, Route, RouteTable, VirtualNetworkGateway, VirtualNetworkGatewayConnection, VirtualNetworkPeering, WebApplicationFirewallPolicy, VirtualDesktopHostPool, VirtualDesktopApplicationGroup, LogicAppWorkflow, LogicAppTriggerCustom, LogicAppActionCustom, ContainerRegistry, ContainerRegistryWebhook, StorageAccount, StorageQueue, StorageShare, StorageTable, StorageBlob, MariadbConfiguration, MariadbDatabase, MariadbFirewallRule, MariadbServer, MariadbVirtualNetworkRule, MysqlConfiguration, MysqlDatabase, MysqlFirewallRule, MysqlServer, MysqlVirtualNetworkRule, PostgresqlConfiguration, PostgresqlDatabase, PostgresqlFirewallRule, PostgresqlServer, PostgresqlVirtualNetworkRule, SQLElasticPool, SQLDatabase, SQLFirewallRule, SQLServer}

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
	_ResourceTypeName[301:337]:        NetworkDdosProtectionPlan,
	_ResourceTypeLowerName[301:337]:   NetworkDdosProtectionPlan,
	_ResourceTypeName[337:353]:        Firewall,
	_ResourceTypeLowerName[337:353]:   Firewall,
	_ResourceTypeName[353:382]:        LocalNetworkGateway,
	_ResourceTypeLowerName[353:382]:   LocalNetworkGateway,
	_ResourceTypeName[382:401]:        NatGateway,
	_ResourceTypeLowerName[382:401]:   NatGateway,
	_ResourceTypeName[401:424]:        NetworkProfile,
	_ResourceTypeLowerName[401:424]:   NetworkProfile,
	_ResourceTypeName[424:453]:        NetworkSecurityRule,
	_ResourceTypeLowerName[424:453]:   NetworkSecurityRule,
	_ResourceTypeName[453:470]:        PublicIP,
	_ResourceTypeLowerName[453:470]:   PublicIP,
	_ResourceTypeName[470:494]:        PublicIPPrefix,
	_ResourceTypeLowerName[470:494]:   PublicIPPrefix,
	_ResourceTypeName[494:507]:        Route,
	_ResourceTypeLowerName[494:507]:   Route,
	_ResourceTypeName[507:526]:        RouteTable,
	_ResourceTypeLowerName[507:526]:   RouteTable,
	_ResourceTypeName[526:557]:        VirtualNetworkGateway,
	_ResourceTypeLowerName[526:557]:   VirtualNetworkGateway,
	_ResourceTypeName[557:599]:        VirtualNetworkGatewayConnection,
	_ResourceTypeLowerName[557:599]:   VirtualNetworkGatewayConnection,
	_ResourceTypeName[599:630]:        VirtualNetworkPeering,
	_ResourceTypeLowerName[599:630]:   VirtualNetworkPeering,
	_ResourceTypeName[630:669]:        WebApplicationFirewallPolicy,
	_ResourceTypeLowerName[630:669]:   WebApplicationFirewallPolicy,
	_ResourceTypeName[669:702]:        VirtualDesktopHostPool,
	_ResourceTypeLowerName[669:702]:   VirtualDesktopHostPool,
	_ResourceTypeName[702:743]:        VirtualDesktopApplicationGroup,
	_ResourceTypeLowerName[702:743]:   VirtualDesktopApplicationGroup,
	_ResourceTypeName[743:769]:        LogicAppWorkflow,
	_ResourceTypeLowerName[743:769]:   LogicAppWorkflow,
	_ResourceTypeName[769:801]:        LogicAppTriggerCustom,
	_ResourceTypeLowerName[769:801]:   LogicAppTriggerCustom,
	_ResourceTypeName[801:832]:        LogicAppActionCustom,
	_ResourceTypeLowerName[801:832]:   LogicAppActionCustom,
	_ResourceTypeName[832:858]:        ContainerRegistry,
	_ResourceTypeLowerName[832:858]:   ContainerRegistry,
	_ResourceTypeName[858:892]:        ContainerRegistryWebhook,
	_ResourceTypeLowerName[858:892]:   ContainerRegistryWebhook,
	_ResourceTypeName[892:915]:        StorageAccount,
	_ResourceTypeLowerName[892:915]:   StorageAccount,
	_ResourceTypeName[915:936]:        StorageQueue,
	_ResourceTypeLowerName[915:936]:   StorageQueue,
	_ResourceTypeName[936:957]:        StorageShare,
	_ResourceTypeLowerName[936:957]:   StorageShare,
	_ResourceTypeName[957:978]:        StorageTable,
	_ResourceTypeLowerName[957:978]:   StorageTable,
	_ResourceTypeName[978:998]:        StorageBlob,
	_ResourceTypeLowerName[978:998]:   StorageBlob,
	_ResourceTypeName[998:1027]:       MariadbConfiguration,
	_ResourceTypeLowerName[998:1027]:  MariadbConfiguration,
	_ResourceTypeName[1027:1051]:      MariadbDatabase,
	_ResourceTypeLowerName[1027:1051]: MariadbDatabase,
	_ResourceTypeName[1051:1080]:      MariadbFirewallRule,
	_ResourceTypeLowerName[1051:1080]: MariadbFirewallRule,
	_ResourceTypeName[1080:1102]:      MariadbServer,
	_ResourceTypeLowerName[1080:1102]: MariadbServer,
	_ResourceTypeName[1102:1138]:      MariadbVirtualNetworkRule,
	_ResourceTypeLowerName[1102:1138]: MariadbVirtualNetworkRule,
	_ResourceTypeName[1138:1165]:      MysqlConfiguration,
	_ResourceTypeLowerName[1138:1165]: MysqlConfiguration,
	_ResourceTypeName[1165:1187]:      MysqlDatabase,
	_ResourceTypeLowerName[1165:1187]: MysqlDatabase,
	_ResourceTypeName[1187:1214]:      MysqlFirewallRule,
	_ResourceTypeLowerName[1187:1214]: MysqlFirewallRule,
	_ResourceTypeName[1214:1234]:      MysqlServer,
	_ResourceTypeLowerName[1214:1234]: MysqlServer,
	_ResourceTypeName[1234:1268]:      MysqlVirtualNetworkRule,
	_ResourceTypeLowerName[1234:1268]: MysqlVirtualNetworkRule,
	_ResourceTypeName[1268:1300]:      PostgresqlConfiguration,
	_ResourceTypeLowerName[1268:1300]: PostgresqlConfiguration,
	_ResourceTypeName[1300:1327]:      PostgresqlDatabase,
	_ResourceTypeLowerName[1300:1327]: PostgresqlDatabase,
	_ResourceTypeName[1327:1359]:      PostgresqlFirewallRule,
	_ResourceTypeLowerName[1327:1359]: PostgresqlFirewallRule,
	_ResourceTypeName[1359:1384]:      PostgresqlServer,
	_ResourceTypeLowerName[1359:1384]: PostgresqlServer,
	_ResourceTypeName[1384:1423]:      PostgresqlVirtualNetworkRule,
	_ResourceTypeLowerName[1384:1423]: PostgresqlVirtualNetworkRule,
	_ResourceTypeName[1423:1446]:      SQLElasticPool,
	_ResourceTypeLowerName[1423:1446]: SQLElasticPool,
	_ResourceTypeName[1446:1466]:      SQLDatabase,
	_ResourceTypeLowerName[1446:1466]: SQLDatabase,
	_ResourceTypeName[1466:1491]:      SQLFirewallRule,
	_ResourceTypeLowerName[1466:1491]: SQLFirewallRule,
	_ResourceTypeName[1491:1509]:      SQLServer,
	_ResourceTypeLowerName[1491:1509]: SQLServer,
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
	_ResourceTypeName[301:337],
	_ResourceTypeName[337:353],
	_ResourceTypeName[353:382],
	_ResourceTypeName[382:401],
	_ResourceTypeName[401:424],
	_ResourceTypeName[424:453],
	_ResourceTypeName[453:470],
	_ResourceTypeName[470:494],
	_ResourceTypeName[494:507],
	_ResourceTypeName[507:526],
	_ResourceTypeName[526:557],
	_ResourceTypeName[557:599],
	_ResourceTypeName[599:630],
	_ResourceTypeName[630:669],
	_ResourceTypeName[669:702],
	_ResourceTypeName[702:743],
	_ResourceTypeName[743:769],
	_ResourceTypeName[769:801],
	_ResourceTypeName[801:832],
	_ResourceTypeName[832:858],
	_ResourceTypeName[858:892],
	_ResourceTypeName[892:915],
	_ResourceTypeName[915:936],
	_ResourceTypeName[936:957],
	_ResourceTypeName[957:978],
	_ResourceTypeName[978:998],
	_ResourceTypeName[998:1027],
	_ResourceTypeName[1027:1051],
	_ResourceTypeName[1051:1080],
	_ResourceTypeName[1080:1102],
	_ResourceTypeName[1102:1138],
	_ResourceTypeName[1138:1165],
	_ResourceTypeName[1165:1187],
	_ResourceTypeName[1187:1214],
	_ResourceTypeName[1214:1234],
	_ResourceTypeName[1234:1268],
	_ResourceTypeName[1268:1300],
	_ResourceTypeName[1300:1327],
	_ResourceTypeName[1327:1359],
	_ResourceTypeName[1359:1384],
	_ResourceTypeName[1384:1423],
	_ResourceTypeName[1423:1446],
	_ResourceTypeName[1446:1466],
	_ResourceTypeName[1466:1491],
	_ResourceTypeName[1491:1509],
}

// ResourceTypeString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func ResourceTypeString(s string) (ResourceType, error) {
	if val, ok := _ResourceTypeNameToValueMap[s]; ok {
		return val, nil
	}

	if val, ok := _ResourceTypeNameToValueMap[strings.ToLower(s)]; ok {
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
