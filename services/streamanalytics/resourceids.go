package streamanalytics

//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=Function -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.StreamAnalytics/streamingjobs/streamingJob1/functions/function1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=StreamingJob -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.StreamAnalytics/streamingjobs/streamingJob1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=StreamingJobSchedule -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.StreamAnalytics/streamingjobs/streamingJob1/schedule/default
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=StreamInput -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.StreamAnalytics/streamingjobs/streamingJob1/inputs/streamInput1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=Output -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.StreamAnalytics/streamingjobs/streamingJob1/outputs/output1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=Cluster -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.StreamAnalytics/clusters/cluster1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=PrivateEndpoint -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.StreamAnalytics/clusters/cluster1/privateEndpoints/endpoint1
