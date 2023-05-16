package notificationhub

//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=Namespace -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.NotificationHubs/namespaces/namespace1 -rewrite=true
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=NotificationHub -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.NotificationHubs/namespaces/namespace1/notificationHubs/hub1 -rewrite=true
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=NotificationHubAuthorizationRule -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.NotificationHubs/namespaces/namespace1/notificationHubs/hub1/authorizationRules/authorizationRule1 -rewrite=true
