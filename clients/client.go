package clients

import (
	"context"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/validation"
	"github.com/hashicorp/terraform-provider-azurerm/common"
	"github.com/hashicorp/terraform-provider-azurerm/features"
	aadb2c "github.com/hashicorp/terraform-provider-azurerm/services/aadb2c/client"
	advisor "github.com/hashicorp/terraform-provider-azurerm/services/advisor/client"
	analysisServices "github.com/hashicorp/terraform-provider-azurerm/services/analysisservices/client"
	apiManagement "github.com/hashicorp/terraform-provider-azurerm/services/apimanagement/client"
	appConfiguration "github.com/hashicorp/terraform-provider-azurerm/services/appconfiguration/client"
	applicationInsights "github.com/hashicorp/terraform-provider-azurerm/services/applicationinsights/client"
	appService "github.com/hashicorp/terraform-provider-azurerm/services/appservice/client"
	attestation "github.com/hashicorp/terraform-provider-azurerm/services/attestation/client"
	authorization "github.com/hashicorp/terraform-provider-azurerm/services/authorization/client"
	automation "github.com/hashicorp/terraform-provider-azurerm/services/automation/client"
	azureStackHCI "github.com/hashicorp/terraform-provider-azurerm/services/azurestackhci/client"
	batch "github.com/hashicorp/terraform-provider-azurerm/services/batch/client"
	blueprints "github.com/hashicorp/terraform-provider-azurerm/services/blueprints/client"
	bot "github.com/hashicorp/terraform-provider-azurerm/services/bot/client"
	cdn "github.com/hashicorp/terraform-provider-azurerm/services/cdn/client"
	cognitiveServices "github.com/hashicorp/terraform-provider-azurerm/services/cognitive/client"
	communication "github.com/hashicorp/terraform-provider-azurerm/services/communication/client"
	compute "github.com/hashicorp/terraform-provider-azurerm/services/compute/client"
	confidentialledger "github.com/hashicorp/terraform-provider-azurerm/services/confidentialledger/client"
	connections "github.com/hashicorp/terraform-provider-azurerm/services/connections/client"
	consumption "github.com/hashicorp/terraform-provider-azurerm/services/consumption/client"
	containerServices "github.com/hashicorp/terraform-provider-azurerm/services/containers/client"
	cosmosdb "github.com/hashicorp/terraform-provider-azurerm/services/cosmos/client"
	costmanagement "github.com/hashicorp/terraform-provider-azurerm/services/costmanagement/client"
	customproviders "github.com/hashicorp/terraform-provider-azurerm/services/customproviders/client"
	datamigration "github.com/hashicorp/terraform-provider-azurerm/services/databasemigration/client"
	databoxedge "github.com/hashicorp/terraform-provider-azurerm/services/databoxedge/client"
	databricks "github.com/hashicorp/terraform-provider-azurerm/services/databricks/client"
	datafactory "github.com/hashicorp/terraform-provider-azurerm/services/datafactory/client"
	dataprotection "github.com/hashicorp/terraform-provider-azurerm/services/dataprotection/client"
	datashare "github.com/hashicorp/terraform-provider-azurerm/services/datashare/client"
	desktopvirtualization "github.com/hashicorp/terraform-provider-azurerm/services/desktopvirtualization/client"
	devtestlabs "github.com/hashicorp/terraform-provider-azurerm/services/devtestlabs/client"
	digitaltwins "github.com/hashicorp/terraform-provider-azurerm/services/digitaltwins/client"
	disks "github.com/hashicorp/terraform-provider-azurerm/services/disks/client"
	dns "github.com/hashicorp/terraform-provider-azurerm/services/dns/client"
	domainservices "github.com/hashicorp/terraform-provider-azurerm/services/domainservices/client"
	elastic "github.com/hashicorp/terraform-provider-azurerm/services/elastic/client"
	eventgrid "github.com/hashicorp/terraform-provider-azurerm/services/eventgrid/client"
	eventhub "github.com/hashicorp/terraform-provider-azurerm/services/eventhub/client"
	firewall "github.com/hashicorp/terraform-provider-azurerm/services/firewall/client"
	frontdoor "github.com/hashicorp/terraform-provider-azurerm/services/frontdoor/client"
	hdinsight "github.com/hashicorp/terraform-provider-azurerm/services/hdinsight/client"
	healthcare "github.com/hashicorp/terraform-provider-azurerm/services/healthcare/client"
	hpccache "github.com/hashicorp/terraform-provider-azurerm/services/hpccache/client"
	hsm "github.com/hashicorp/terraform-provider-azurerm/services/hsm/client"
	iotcentral "github.com/hashicorp/terraform-provider-azurerm/services/iotcentral/client"
	iothub "github.com/hashicorp/terraform-provider-azurerm/services/iothub/client"
	timeseriesinsights "github.com/hashicorp/terraform-provider-azurerm/services/iottimeseriesinsights/client"
	keyvault "github.com/hashicorp/terraform-provider-azurerm/services/keyvault/client"
	kusto "github.com/hashicorp/terraform-provider-azurerm/services/kusto/client"
	legacy "github.com/hashicorp/terraform-provider-azurerm/services/legacy/client"
	lighthouse "github.com/hashicorp/terraform-provider-azurerm/services/lighthouse/client"
	loadbalancers "github.com/hashicorp/terraform-provider-azurerm/services/loadbalancer/client"
	loadtest "github.com/hashicorp/terraform-provider-azurerm/services/loadtest/client"
	loganalytics "github.com/hashicorp/terraform-provider-azurerm/services/loganalytics/client"
	logic "github.com/hashicorp/terraform-provider-azurerm/services/logic/client"
	logz "github.com/hashicorp/terraform-provider-azurerm/services/logz/client"
	machinelearning "github.com/hashicorp/terraform-provider-azurerm/services/machinelearning/client"
	maintenance "github.com/hashicorp/terraform-provider-azurerm/services/maintenance/client"
	managedapplication "github.com/hashicorp/terraform-provider-azurerm/services/managedapplications/client"
	managementgroup "github.com/hashicorp/terraform-provider-azurerm/services/managementgroup/client"
	maps "github.com/hashicorp/terraform-provider-azurerm/services/maps/client"
	mariadb "github.com/hashicorp/terraform-provider-azurerm/services/mariadb/client"
	media "github.com/hashicorp/terraform-provider-azurerm/services/media/client"
	mixedreality "github.com/hashicorp/terraform-provider-azurerm/services/mixedreality/client"
	monitor "github.com/hashicorp/terraform-provider-azurerm/services/monitor/client"
	msi "github.com/hashicorp/terraform-provider-azurerm/services/msi/client"
	mssql "github.com/hashicorp/terraform-provider-azurerm/services/mssql/client"
	mysql "github.com/hashicorp/terraform-provider-azurerm/services/mysql/client"
	netapp "github.com/hashicorp/terraform-provider-azurerm/services/netapp/client"
	network "github.com/hashicorp/terraform-provider-azurerm/services/network/client"
	notificationhub "github.com/hashicorp/terraform-provider-azurerm/services/notificationhub/client"
	policy "github.com/hashicorp/terraform-provider-azurerm/services/policy/client"
	portal "github.com/hashicorp/terraform-provider-azurerm/services/portal/client"
	postgres "github.com/hashicorp/terraform-provider-azurerm/services/postgres/client"
	powerBI "github.com/hashicorp/terraform-provider-azurerm/services/powerbi/client"
	privatedns "github.com/hashicorp/terraform-provider-azurerm/services/privatedns/client"
	purview "github.com/hashicorp/terraform-provider-azurerm/services/purview/client"
	recoveryServices "github.com/hashicorp/terraform-provider-azurerm/services/recoveryservices/client"
	redis "github.com/hashicorp/terraform-provider-azurerm/services/redis/client"
	redisenterprise "github.com/hashicorp/terraform-provider-azurerm/services/redisenterprise/client"
	relay "github.com/hashicorp/terraform-provider-azurerm/services/relay/client"
	resource "github.com/hashicorp/terraform-provider-azurerm/services/resource/client"
	search "github.com/hashicorp/terraform-provider-azurerm/services/search/client"
	securityCenter "github.com/hashicorp/terraform-provider-azurerm/services/securitycenter/client"
	sentinel "github.com/hashicorp/terraform-provider-azurerm/services/sentinel/client"
	serviceBus "github.com/hashicorp/terraform-provider-azurerm/services/servicebus/client"
	serviceFabric "github.com/hashicorp/terraform-provider-azurerm/services/servicefabric/client"
	serviceFabricManaged "github.com/hashicorp/terraform-provider-azurerm/services/servicefabricmanaged/client"
	signalr "github.com/hashicorp/terraform-provider-azurerm/services/signalr/client"
	appPlatform "github.com/hashicorp/terraform-provider-azurerm/services/springcloud/client"
	sql "github.com/hashicorp/terraform-provider-azurerm/services/sql/client"
	storage "github.com/hashicorp/terraform-provider-azurerm/services/storage/client"
	streamAnalytics "github.com/hashicorp/terraform-provider-azurerm/services/streamanalytics/client"
	subscription "github.com/hashicorp/terraform-provider-azurerm/services/subscription/client"
	synapse "github.com/hashicorp/terraform-provider-azurerm/services/synapse/client"
	trafficManager "github.com/hashicorp/terraform-provider-azurerm/services/trafficmanager/client"
	videoAnalyzer "github.com/hashicorp/terraform-provider-azurerm/services/videoanalyzer/client"
	vmware "github.com/hashicorp/terraform-provider-azurerm/services/vmware/client"
	web "github.com/hashicorp/terraform-provider-azurerm/services/web/client"
)

type Client struct {
	// StopContext is used for propagating control from Terraform Core (e.g. Ctrl/Cmd+C)
	StopContext context.Context

	Account  *ResourceManagerAccount
	Features features.UserFeatures

	AadB2c                *aadb2c.Client
	Advisor               *advisor.Client
	AnalysisServices      *analysisServices.Client
	ApiManagement         *apiManagement.Client
	AppConfiguration      *appConfiguration.Client
	AppInsights           *applicationInsights.Client
	AppPlatform           *appPlatform.Client
	AppService            *appService.Client
	Attestation           *attestation.Client
	Authorization         *authorization.Client
	Automation            *automation.Client
	AzureStackHCI         *azureStackHCI.Client
	Batch                 *batch.Client
	Blueprints            *blueprints.Client
	Bot                   *bot.Client
	Cdn                   *cdn.Client
	Cognitive             *cognitiveServices.Client
	Communication         *communication.Client
	Compute               *compute.Client
	ConfidentialLedger    *confidentialledger.Client
	Connections           *connections.Client
	Consumption           *consumption.Client
	Containers            *containerServices.Client
	Cosmos                *cosmosdb.Client
	CostManagement        *costmanagement.Client
	CustomProviders       *customproviders.Client
	DatabaseMigration     *datamigration.Client
	DataBricks            *databricks.Client
	DataboxEdge           *databoxedge.Client
	DataFactory           *datafactory.Client
	DataProtection        *dataprotection.Client
	DataShare             *datashare.Client
	DesktopVirtualization *desktopvirtualization.Client
	DevTestLabs           *devtestlabs.Client
	DigitalTwins          *digitaltwins.Client
	Disks                 *disks.Client
	Dns                   *dns.Client
	DomainServices        *domainservices.Client
	Elastic               *elastic.Client
	EventGrid             *eventgrid.Client
	Eventhub              *eventhub.Client
	Firewall              *firewall.Client
	Frontdoor             *frontdoor.Client
	HPCCache              *hpccache.Client
	HSM                   *hsm.Client
	HDInsight             *hdinsight.Client
	HealthCare            *healthcare.Client
	IoTCentral            *iotcentral.Client
	IoTHub                *iothub.Client
	IoTTimeSeriesInsights *timeseriesinsights.Client
	KeyVault              *keyvault.Client
	Kusto                 *kusto.Client
	Legacy                *legacy.Client
	Lighthouse            *lighthouse.Client
	LoadBalancers         *loadbalancers.Client
	LoadTest              *loadtest.Client
	LogAnalytics          *loganalytics.Client
	Logic                 *logic.Client
	Logz                  *logz.Client
	MachineLearning       *machinelearning.Client
	Maintenance           *maintenance.Client
	ManagedApplication    *managedapplication.Client
	ManagementGroups      *managementgroup.Client
	Maps                  *maps.Client
	MariaDB               *mariadb.Client
	Media                 *media.Client
	MixedReality          *mixedreality.Client
	Monitor               *monitor.Client
	MSI                   *msi.Client
	MSSQL                 *mssql.Client
	MySQL                 *mysql.Client
	NetApp                *netapp.Client
	Network               *network.Client
	NotificationHubs      *notificationhub.Client
	Policy                *policy.Client
	Portal                *portal.Client
	Postgres              *postgres.Client
	PowerBI               *powerBI.Client
	PrivateDns            *privatedns.Client
	Purview               *purview.Client
	RecoveryServices      *recoveryServices.Client
	Redis                 *redis.Client
	RedisEnterprise       *redisenterprise.Client
	Relay                 *relay.Client
	Resource              *resource.Client
	Search                *search.Client
	SecurityCenter        *securityCenter.Client
	Sentinel              *sentinel.Client
	ServiceBus            *serviceBus.Client
	ServiceFabric         *serviceFabric.Client
	ServiceFabricManaged  *serviceFabricManaged.Client
	SignalR               *signalr.Client
	Storage               *storage.Client
	StreamAnalytics       *streamAnalytics.Client
	Subscription          *subscription.Client
	Sql                   *sql.Client
	Synapse               *synapse.Client
	TrafficManager        *trafficManager.Client
	VideoAnalyzer         *videoAnalyzer.Client
	Vmware                *vmware.Client
	Web                   *web.Client
}

// NOTE: it should be possible for this method to become Private once the top level Client's removed

func (client *Client) Build(ctx context.Context, o *common.ClientOptions) error {
	autorest.Count429AsRetry = false
	// Disable the Azure SDK for Go's validation since it's unhelpful for our use-case
	validation.Disabled = true

	client.Features = o.Features
	client.StopContext = ctx

	client.AadB2c = aadb2c.NewClient(o)
	client.Advisor = advisor.NewClient(o)
	client.AnalysisServices = analysisServices.NewClient(o)
	client.ApiManagement = apiManagement.NewClient(o)
	client.AppConfiguration = appConfiguration.NewClient(o)
	client.AppInsights = applicationInsights.NewClient(o)
	client.AppPlatform = appPlatform.NewClient(o)
	client.AppService = appService.NewClient(o)
	client.Attestation = attestation.NewClient(o)
	client.Authorization = authorization.NewClient(o)
	client.Automation = automation.NewClient(o)
	client.AzureStackHCI = azureStackHCI.NewClient(o)
	client.Batch = batch.NewClient(o)
	client.Blueprints = blueprints.NewClient(o)
	client.Bot = bot.NewClient(o)
	client.Cdn = cdn.NewClient(o)
	client.Cognitive = cognitiveServices.NewClient(o)
	client.Communication = communication.NewClient(o)
	client.Compute = compute.NewClient(o)
	client.ConfidentialLedger = confidentialledger.NewClient(o)
	client.Connections = connections.NewClient(o)
	client.Consumption = consumption.NewClient(o)
	client.Containers = containerServices.NewClient(o)
	client.Cosmos = cosmosdb.NewClient(o)
	client.CostManagement = costmanagement.NewClient(o)
	client.CustomProviders = customproviders.NewClient(o)
	client.DatabaseMigration = datamigration.NewClient(o)
	client.DataBricks = databricks.NewClient(o)
	client.DataboxEdge = databoxedge.NewClient(o)
	client.DataFactory = datafactory.NewClient(o)
	client.DataProtection = dataprotection.NewClient(o)
	client.DataShare = datashare.NewClient(o)
	client.DesktopVirtualization = desktopvirtualization.NewClient(o)
	client.DevTestLabs = devtestlabs.NewClient(o)
	client.DigitalTwins = digitaltwins.NewClient(o)
	client.Disks = disks.NewClient(o)
	client.Dns = dns.NewClient(o)
	client.DomainServices = domainservices.NewClient(o)
	client.Elastic = elastic.NewClient(o)
	client.EventGrid = eventgrid.NewClient(o)
	client.Eventhub = eventhub.NewClient(o)
	client.Firewall = firewall.NewClient(o)
	client.Frontdoor = frontdoor.NewClient(o)
	client.HPCCache = hpccache.NewClient(o)
	client.HSM = hsm.NewClient(o)
	client.HDInsight = hdinsight.NewClient(o)
	client.HealthCare = healthcare.NewClient(o)
	client.IoTCentral = iotcentral.NewClient(o)
	client.IoTHub = iothub.NewClient(o)
	client.IoTTimeSeriesInsights = timeseriesinsights.NewClient(o)
	client.KeyVault = keyvault.NewClient(o)
	client.Kusto = kusto.NewClient(o)
	client.Legacy = legacy.NewClient(o)
	client.Lighthouse = lighthouse.NewClient(o)
	client.LogAnalytics = loganalytics.NewClient(o)
	client.LoadBalancers = loadbalancers.NewClient(o)
	client.LoadTest = loadtest.NewClient(o)
	client.Logic = logic.NewClient(o)
	client.Logz = logz.NewClient(o)
	client.MachineLearning = machinelearning.NewClient(o)
	client.Maintenance = maintenance.NewClient(o)
	client.ManagedApplication = managedapplication.NewClient(o)
	client.ManagementGroups = managementgroup.NewClient(o)
	client.Maps = maps.NewClient(o)
	client.MariaDB = mariadb.NewClient(o)
	client.Media = media.NewClient(o)
	client.MixedReality = mixedreality.NewClient(o)
	client.Monitor = monitor.NewClient(o)
	client.MSI = msi.NewClient(o)
	client.MSSQL = mssql.NewClient(o)
	client.MySQL = mysql.NewClient(o)
	client.NetApp = netapp.NewClient(o)
	client.Network = network.NewClient(o)
	client.NotificationHubs = notificationhub.NewClient(o)
	client.Policy = policy.NewClient(o)
	client.Portal = portal.NewClient(o)
	client.Postgres = postgres.NewClient(o)
	client.PowerBI = powerBI.NewClient(o)
	client.PrivateDns = privatedns.NewClient(o)
	client.Purview = purview.NewClient(o)
	client.RecoveryServices = recoveryServices.NewClient(o)
	client.Redis = redis.NewClient(o)
	client.RedisEnterprise = redisenterprise.NewClient(o)
	client.Relay = relay.NewClient(o)
	client.Resource = resource.NewClient(o)
	client.Search = search.NewClient(o)
	client.SecurityCenter = securityCenter.NewClient(o)
	client.Sentinel = sentinel.NewClient(o)
	client.ServiceBus = serviceBus.NewClient(o)
	client.ServiceFabric = serviceFabric.NewClient(o)
	client.ServiceFabricManaged = serviceFabricManaged.NewClient(o)
	client.SignalR = signalr.NewClient(o)
	client.Sql = sql.NewClient(o)
	client.Storage = storage.NewClient(o)
	client.StreamAnalytics = streamAnalytics.NewClient(o)
	client.Subscription = subscription.NewClient(o)
	client.Synapse = synapse.NewClient(o)
	client.TrafficManager = trafficManager.NewClient(o)
	client.VideoAnalyzer = videoAnalyzer.NewClient(o)
	client.Vmware = vmware.NewClient(o)
	client.Web = web.NewClient(o)

	return nil
}
