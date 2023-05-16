package loganalytics

//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=DataSource -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.OperationalInsights/workspaces/workspace1/dataSources/dataSource1 -rewrite=true

//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=LogAnalyticsCluster -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.OperationalInsights/clusters/cluster1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=LogAnalyticsDataExport -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.OperationalInsights/workspaces/workspace1/dataexports/dataExport1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=LogAnalyticsLinkedService -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.OperationalInsights/workspaces/workspace1/linkedServices/linkedService1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=LogAnalyticsLinkedStorageAccount -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.OperationalInsights/workspaces/workspace1/linkedStorageAccounts/query
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=LogAnalyticsSavedSearch -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.OperationalInsights/workspaces/workspace1/savedSearches/search1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=LogAnalyticsSolution -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.OperationsManagement/solutions/solution1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=LogAnalyticsStorageInsights -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.OperationalInsights/workspaces/workspace1/storageInsightConfigs/storageInsight1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=LogAnalyticsWorkspace -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.OperationalInsights/workspaces/workspace1
