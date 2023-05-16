package appservice

import (
	"github.com/Azure/azure-sdk-for-go/services/web/mgmt/2021-02-01/web"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type GithubActionConfiguration struct {
	CodeConfig           []GitHubActionCodeConfig      `tfschema:"code_configuration"`
	ContainerConfig      []GitHubActionContainerConfig `tfschema:"container_configuration"`
	UsesLinux            bool                          `tfschema:"linux_action"`
	GenerateWorkflowFile bool                          `tfschema:"generate_workflow_file"`
}

type GitHubActionCodeConfig struct {
	RuntimeStack   string `tfschema:"runtime_stack"`
	RuntimeVersion string `tfschema:"runtime_version"`
}

type GitHubActionContainerConfig struct { // TODO - May need to remove this, cannot find a valid way to use it?
	RegistryURL      string `tfschema:"registry_url"`
	ImageName        string `tfschema:"image_name"`
	RegistryUsername string `tfschema:"registry_username"`
	RegistryPassword string `tfschema:"registry_password"`
}

func githubActionConfigSchema() *pluginsdk.Schema {
	return &pluginsdk.Schema{
		Type:     pluginsdk.TypeList,
		MaxItems: 1,
		ForceNew: true,
		Optional: true,
		Elem: &pluginsdk.Resource{
			Schema: map[string]*pluginsdk.Schema{
				"container_configuration": {
					Type:     pluginsdk.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &pluginsdk.Resource{
						Schema: map[string]*pluginsdk.Schema{
							"registry_url": {
								Type:        pluginsdk.TypeString,
								Required:    true,
								ForceNew:    true,
								Description: "The server URL for the container registry where the build will be hosted.",
								// ValidateFunc: validation.IsURLWithHTTPorHTTPS,
							},

							"image_name": {
								Type:        pluginsdk.TypeString,
								Required:    true,
								ForceNew:    true,
								Description: "The image name for the build.",
								// ValidateFunc: validate.NoEmptyStrings,
							},

							"registry_username": {
								Type:        pluginsdk.TypeString,
								Optional:    true,
								ForceNew:    true,
								Description: "The username used to upload the image to the container registry.",
							},

							"registry_password": {
								Type:        pluginsdk.TypeString,
								Optional:    true,
								ForceNew:    true,
								Sensitive:   true,
								Description: "The password used to upload the image to the container registry.",
							},
						},
					},
				},

				"code_configuration": {
					Type:     pluginsdk.TypeList,
					Optional: true,
					MaxItems: 1,
					ForceNew: true,
					Elem: &pluginsdk.Resource{
						Schema: map[string]*pluginsdk.Schema{
							"runtime_stack": {
								Type:     pluginsdk.TypeString,
								Required: true,
								ForceNew: true,
								ValidateFunc: validation.StringInSlice([]string{
									"dotnetcore", // For all .Net builds on all O/S
									"spring",     // Java on Java, JBOSSEAP
									"tomcat",     // Java on Tomcat
									"node",       // Node, all versions
									"python",     // Python, all versions
								}, false),
								Description: "The value to use for the Runtime Stack in the workflow file content for code base apps.",
							},

							"runtime_version": {
								Type:         pluginsdk.TypeString,
								Required:     true,
								ForceNew:     true,
								ValidateFunc: validation.StringIsNotEmpty,
								Description:  "The value to use for the Runtime Version in the workflow file content for code base apps.",
							},
						},
					},
				},

				"generate_workflow_file": {
					Type:        pluginsdk.TypeBool,
					Optional:    true,
					ForceNew:    true,
					Default:     true,
					Description: "Should the service generate the GitHub Action Workflow file. Defaults to `true`",
				},

				"linux_action": {
					Type:        pluginsdk.TypeBool,
					Computed:    true,
					Description: "Denotes this action uses a Linux base image.",
				},
			},
		},
	}
}

func expandGithubActionConfig(input []GithubActionConfiguration, usesLinux bool) *web.GitHubActionConfiguration {
	if len(input) == 0 {
		return nil
	}

	ghActionConfig := input[0]
	output := &web.GitHubActionConfiguration{
		CodeConfiguration:      nil,
		ContainerConfiguration: nil,
		IsLinux:                utils.Bool(usesLinux),
		GenerateWorkflowFile:   utils.Bool(ghActionConfig.GenerateWorkflowFile),
	}

	if len(ghActionConfig.CodeConfig) != 0 {
		codeConfig := ghActionConfig.CodeConfig[0]
		output.CodeConfiguration = &web.GitHubActionCodeConfiguration{
			RuntimeStack:   utils.String(codeConfig.RuntimeStack),
			RuntimeVersion: utils.String(codeConfig.RuntimeVersion),
		}
	}

	if len(ghActionConfig.ContainerConfig) != 0 {
		containerConfig := ghActionConfig.ContainerConfig[0]
		output.ContainerConfiguration = &web.GitHubActionContainerConfiguration{
			ServerURL: utils.String(containerConfig.RegistryURL),
			ImageName: utils.String(containerConfig.ImageName),
			Username:  utils.String(containerConfig.RegistryUsername),
			Password:  utils.String(containerConfig.RegistryPassword),
		}
	}

	return output
}

func flattenGitHubActionConfiguration(input *web.GitHubActionConfiguration) []GithubActionConfiguration {
	output := make([]GithubActionConfiguration, 0)
	if input == nil {
		return output
	}

	isLinux := false
	if v := input.IsLinux; v != nil {
		isLinux = *v
	}
	genWorkflow := false
	if v := input.GenerateWorkflowFile; v != nil {
		genWorkflow = *v
	}
	ghConfig := GithubActionConfiguration{
		UsesLinux:            isLinux,
		GenerateWorkflowFile: genWorkflow,
	}

	if codeConfig := input.CodeConfiguration; codeConfig != nil {
		ghCodeConfig := []GitHubActionCodeConfig{{
			RuntimeStack:   utils.NormalizeNilableString(codeConfig.RuntimeStack),
			RuntimeVersion: utils.NormalizeNilableString(codeConfig.RuntimeVersion),
		}}
		ghConfig.CodeConfig = ghCodeConfig
	}

	if containerConfig := input.ContainerConfiguration; containerConfig != nil {
		ghContainerConfig := []GitHubActionContainerConfig{{
			RegistryPassword: utils.NormalizeNilableString(containerConfig.Password),
			RegistryUsername: utils.NormalizeNilableString(containerConfig.Username),
			RegistryURL:      utils.NormalizeNilableString(containerConfig.ServerURL),
			ImageName:        utils.NormalizeNilableString(containerConfig.ImageName),
		}}
		ghConfig.ContainerConfig = ghContainerConfig
	}

	output = append(output, ghConfig)

	return output
}
