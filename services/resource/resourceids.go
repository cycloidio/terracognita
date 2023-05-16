package resource

//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=ResourceGroup -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/group1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -rewrite=true -name=ResourceGroupTemplateDeployment -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/group1/providers/Microsoft.Resources/deployments/deploy1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=SubscriptionTemplateDeployment -id=/subscriptions/12345678-1234-9876-4563-123456789012/providers/Microsoft.Resources/deployments/deploy1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=TemplateSpecVersion -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/templateSpecRG/providers/Microsoft.Resources/templateSpecs/templateSpec1/versions/v1.0

// ResourceProvider is manually maintained since the generator doesn't support outputting this information at this time
