package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

var functions = []Function{
	// cloud dns
	Function{Resource: "ManagedZone", API: "dns", AddAPISufix: true, ResourceList: "ManagedZonesListResponse", NoFilter: true, ItemName: "ManagedZones"},
	Function{Resource: "Policy", API: "dns", AddAPISufix: true, ResourceList: "PoliciesListResponse", NoFilter: true, ItemName: "Policies"},
	Function{Resource: "ResourceRecordSet", API: "dns", AddAPISufix: true, OtherListArg: "managedZones", ResourceList: "ResourceRecordSetsListResponse", NoFilter: true, ItemName: "Rrsets"},
	// cloud platform
	Function{Resource: "BillingAccount", FunctionName: "ListBillingSubaccounts", NoProjectScope: true, MaxResultFunc: "PageSize", API: "cloudbilling", ResourceList: "ListBillingAccountsResponse", NoFilter: true, ItemName: "BillingAccounts"},
	Function{Resource: "Role", FunctionName: "ListProjectIAMCustomRoles", NoProjectScope: true, ParentFunction: true, MaxResultFunc: "PageSize", API: "iam", ResourceList: "ListRolesResponse", NoFilter: true, ItemName: "Roles"},
	// cloud sql
	Function{Resource: "DatabaseInstance", FunctionName: "ListSQLDatabaseInstances", PluralName: "StorageInstances", API: "sqladmin", ResourceList: "InstancesListResponse", ServiceName: "Instances"},
	// cloud storage
	Function{Resource: "Bucket", NoFilter: true, AddAPISufix: true, API: "storage", ResourceList: "Buckets"},
	// compute
	Function{Resource: "Address", Region: true},
	Function{Resource: "Autoscaler", Zone: true},
	Function{Resource: "BackendService"},
	Function{Resource: "BackendBucket"},
	Function{Resource: "Disk", Zone: true},
	Function{Resource: "Firewall"},
	Function{Resource: "ForwardingRule", PluralName: "GlobalForwardingRules", ServiceName: "GlobalForwardingRules"},
	Function{Resource: "ForwardingRule", Region: true},
	Function{Resource: "HealthCheck"},
	Function{Resource: "Instance", Zone: true},
	Function{Resource: "InstanceGroup", Zone: true},
	Function{Resource: "Network"},
	Function{Resource: "SslCertificate", PluralName: "SSLCertificates"},
	Function{Resource: "TargetHttpProxy", PluralName: "TargetHTTPProxies", ServiceName: "TargetHttpProxies"},
	Function{Resource: "TargetHttpsProxy", PluralName: "TargetHTTPSProxies", ServiceName: "TargetHttpsProxies"},
	Function{Resource: "UrlMap", PluralName: "URLMaps"},
	Function{Resource: "Address", Region: true, FunctionName: "ListGlobalAddresses", PluralName: "GlobalAddress", ResourceList: "AddressList"},
	Function{Resource: "Image"},
	Function{Resource: "InstanceGroupManager", Zone: true},
	Function{Resource: "InstanceTemplate"},
	Function{Resource: "SslCertificate", FunctionName: "ListManagedSslCertificates"},
	Function{Resource: "NetworkEndpointGroup", Zone: true},
	Function{Resource: "Route"},
	Function{Resource: "SecurityPolicy"},
	Function{Resource: "ServiceAttachment", Region: true},
	Function{Resource: "Snapshot"},
	Function{Resource: "SslPolicy", ResourceList: "SslPoliciesList"},
	Function{Resource: "Subnetwork", Region: true},
	Function{Resource: "TargetGrpcProxy"},
	Function{Resource: "TargetInstance", Zone: true},
	Function{Resource: "TargetPool", Region: true},
	Function{Resource: "TargetSslProxy"},
	Function{Resource: "TargetTcpProxy", FunctionName: "ListTargetTCPProxies"},
	//file
	Function{Resource: "Instance", FunctionName: "ListFilestoreInstances", API: "file", ServiceName: "ProjectsLocationsInstances", MaxResultFunc: "PageSize", ParentListScope: true, ResourceList: "ListInstancesResponse", ItemName: "Instances"},
	// kubernetes container engine
	Function{Resource: "Cluster", AddAPISufix: true, ServiceName: "ProjectsLocationsClusters", API: "container", DoMethodToList: true, ParentListScope: true, NoProjectScope: true, ItemName: "Clusters"},
	//	redis
	Function{Resource: "Instance", FunctionName: "ListRedisInstances", API: "redis", ServiceName: "ProjectsLocationsInstances", MaxResultFunc: "PageSize", NoFilter: true, ParentListScope: true, ResourceList: "ListInstancesResponse", ItemName: "Instances"},
	//	logging
	Function{Resource: "LogMetric", ServiceName: "ProjectsMetrics", API: "logging", MaxResultFunc: "PageSize", ParentListScope: true, NoFilter: true, ResourceList: "ListLogMetricsResponse", ItemName: "Metrics"},
	// monitoring
	Function{Resource: "AlertPolicy", API: "monitoring", AddAPISufix: true, ServiceName: "ProjectsAlertPolicies", ParentListScope: true, MaxResultFunc: "PageSize", ResourceList: "ListAlertPoliciesResponse", ItemName: "AlertPolicies"},
	Function{Resource: "Group", API: "monitoring", AddAPISufix: true, ServiceName: "ProjectsGroups", ParentListScope: true, MaxResultFunc: "PageSize", NoFilter: true, ResourceList: "ListGroupsResponse", ItemName: "Group"},
	Function{Resource: "NotificationChannel", API: "monitoring", AddAPISufix: true, ServiceName: "ProjectsNotificationChannels", ParentListScope: true, MaxResultFunc: "PageSize", NoFilter: true, ResourceList: "ListNotificationChannelsResponse", ItemName: "NotificationChannels"},
	Function{Resource: "UptimeCheckConfig", API: "monitoring", AddAPISufix: true, ServiceName: "ProjectsUptimeCheckConfigs", ParentListScope: true, MaxResultFunc: "PageSize", NoFilter: true, ResourceList: "ListUptimeCheckConfigsResponse", ItemName: "UptimeCheckConfigs"},
}

func main() {
	f, err := os.OpenFile("./reader_generated.go", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := generate(f, functions); err != nil {
		panic(err)
	}
}

func generate(opt io.Writer, fns []Function) error {
	var fnBuff = bytes.Buffer{}

	if err := pkgTmpl.Execute(&fnBuff, nil); err != nil {
		return errors.Wrap(err, "unable to execute package template")
	}

	for _, function := range functions {
		if err := function.Execute(&fnBuff); err != nil {
			return errors.Wrapf(err, "unable to execute function template for: %s", function.Resource)
		}
	}

	// format
	cmd := exec.Command("goimports")
	cmd.Stdin = &fnBuff
	cmd.Stdout = opt
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "unable to run goimports command")
	}
	return nil
}
