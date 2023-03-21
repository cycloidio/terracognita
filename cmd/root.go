package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/adrg/xdg"
	"github.com/cycloidio/mxwriter"
	"github.com/cycloidio/terracognita/filter"
	"github.com/cycloidio/terracognita/hcl"
	"github.com/cycloidio/terracognita/log"
	"github.com/cycloidio/terracognita/provider"
	"github.com/cycloidio/terracognita/state"
	"github.com/cycloidio/terracognita/tag"
	"github.com/cycloidio/terracognita/writer"
	kitlog "github.com/go-kit/kit/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

var (
	isHCLDir bool
	noTags   []tag.Tag = nil
	hclOut   io.ReadWriter
	stateOut io.Writer

	closeOut = make([]io.Closer, 0, 0)

	include, exclude, targets []string
	logsOut                   io.Writer

	// RootCmd it's the entry command for the cmd on terracognita
	RootCmd = &cobra.Command{
		Use:   "terracognita",
		Short: "Reads from Providers and generates a Terraform configuration",
		Long:  "Reads from Providers and generates a Terraform configuration, all the flags can be used also with ENV (ex: --aws-access-key == AWS_ACCESS_KEY)",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			err := os.MkdirAll(path.Dir(viper.GetString("log-file")), 0700)
			if err != nil {
				return err
			}
			logFile, err := os.OpenFile(viper.GetString("log-file"), os.O_APPEND|os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0644)
			if err != nil {
				return err
			}
			closeOut = append(closeOut, logFile)
			// Initialize the logs by setting by default the logs
			// to Stdout, but if 'v' or 'd' is defined the logger
			// will be initialized and structured logs will be used
			// and if 'd' it's defined TF_LOG will be used too
			if viper.GetBool("verbose") || viper.GetBool("debug") {
				logsOut = ioutil.Discard
				w := io.MultiWriter(os.Stdout, logFile)
				log.Init(w, viper.GetBool("debug"))
			} else {
				logsOut = os.Stdout
				log.Init(logFile, false)
			}

			return nil
		},
	}
)

func requiredStringFlags(names ...string) error {
	for _, n := range names {
		if viper.GetString(n) == "" {
			return fmt.Errorf("the flag %q is required", n)
		}
	}

	return nil
}

func preRunEOutput(cmd *cobra.Command, args []string) error {
	// Initializes/Validates the HCL and TFSTATE flags
	if module := viper.GetString("module"); module != "" {

		// We check if there is any error checking it
		// if its NotExist we do not care as we'll create it
		// but if it exists and it's not a Dir then we fail
		fi, err := os.Stat(module)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if err == nil && !fi.IsDir() {
			return fmt.Errorf("the --module must be a directory: %s", module)
		}

		// It means is a existing directory
		// so we'll ask for confirmation before deleting the
		// existent content
		if err == nil {
			fmt.Printf("We are about to remove all content from %q, are you sure? Yes/No (Y/N):\n", module)
			var s string
			fmt.Scanf("%s\n", &s)
			s = strings.ToLower(s)
			if s != "yes" && s != "y" {
				return errors.New("the import was stopped")
			}
		}

		// Clean the module dir
		err = os.RemoveAll(module)
		if err != nil {
			return err
		}

		// Recreate it just if it was not created
		// RemoveAll will not return error if
		// it does not exists
		err = os.MkdirAll(module, 0700)
		if err != nil {
			return err
		}

		hclOut = mxwriter.NewMux()
	} else if hcl := viper.GetString("hcl"); hcl != "" {
		// We check if there is any error checking it
		// if its NotExist we do not care as we'll create it
		fi, err := os.Stat(hcl)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		hasExt := filepath.Ext(hcl) != ""
		if (err == nil && !fi.IsDir()) || hasExt {
			isHCLDir = false
		} else {
			isHCLDir = true
			// It means is a existing directory
			// so we'll ask for confirmation before deleting the
			// existent content
			if err == nil {
				fmt.Printf("We are about to remove all content from %q, are you sure? Yes/No (Y/N):\n", hcl)
				var s string
				fmt.Scanf("%s\n", &s)
				s = strings.ToLower(s)
				if s != "yes" && s != "y" {
					return errors.New("the import was stopped")
				}
			}

			// Clean the module dir
			err = os.RemoveAll(hcl)
			if err != nil {
				return err
			}

			// Recreate it just if it was not created
			// RemoveAll will not return error if
			// it does not exists
			err = os.MkdirAll(hcl, 0700)
			if err != nil {
				return err
			}
		}

		hclOut = mxwriter.NewMux()
	}
	if viper.GetString("tfstate") != "" {
		f, err := os.OpenFile(viper.GetString("tfstate"), os.O_APPEND|os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return fmt.Errorf("could not OpenFile %s because: %s", viper.GetString("tfstate"), err)
		}
		stateOut = f
		closeOut = append(closeOut, f)
	}

	if viper.GetString("tfstate") == "" && viper.GetString("hcl") == "" && viper.GetString("module") == "" {
		return fmt.Errorf("one of --module, --hcl  or --tfstate are required")
	}
	return nil
}

func postRunEOutput(cmd *cobra.Command, args []string) error {
	// Closes all the opened files
	for _, c := range closeOut {
		if err := c.Close(); err != nil {
			return err
		}
	}

	if m := viper.GetString("module"); m != "" {
		dm, err := mxwriter.NewDemux(hclOut)
		if err != nil {
			return err
		}

		moduleName := filepath.Base(m)
		mdir := fmt.Sprintf("module-%s", moduleName)

		err = os.Mkdir(filepath.Join(m, mdir), 0700)
		if err != nil {
			return err
		}

		for _, k := range dm.Keys() {
			var (
				filep string
			)
			if k == writer.ModuleCategoryKey {
				filep = filepath.Join(m, "module.tf")
			} else {
				filep = filepath.Join(m, mdir, fmt.Sprintf("%s.tf", k))
			}

			f, err := os.OpenFile(filep, os.O_APPEND|os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0644)
			if err != nil {
				return fmt.Errorf("could not OpenFile %s because: %s", filep, err)
			}
			io.Copy(f, dm.Read(k))
			f.Close()
		}
	} else if hcl := viper.GetString("hcl"); hcl != "" {
		dm, err := mxwriter.NewDemux(hclOut)
		if err != nil {
			return err
		}
		if isHCLDir {
			for _, k := range dm.Keys() {
				filep := filepath.Join(hcl, fmt.Sprintf("%s.tf", k))

				f, err := os.OpenFile(filep, os.O_APPEND|os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0644)
				if err != nil {
					return fmt.Errorf("could not OpenFile %s because: %s", filep, err)
				}
				io.Copy(f, dm.Read(k))
				f.Close()
			}
		} else {
			f, err := os.OpenFile(viper.GetString("hcl"), os.O_APPEND|os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0644)
			if err != nil {
				return fmt.Errorf("could not OpenFile %s because: %s", viper.GetString("hcl"), err)
			}
			io.Copy(f, hclOut)
			f.Close()
		}
	}

	return nil
}

// getWriterOptions will initialize the common writer.Options from the flags
func getWriterOptions() (*writer.Options, error) {
	var module string
	var mv = make(map[string]struct{})
	if m := viper.GetString("module"); m != "" {
		module = filepath.Base(m)

		if pmv := viper.GetString("module-variables"); pmv != "" {
			b, err := ioutil.ReadFile(pmv)
			if err != nil {
				return nil, fmt.Errorf("could not ReadFile on path %q: %w", pmv, err)
			}

			var values map[string][]string
			switch filepath.Ext(pmv) {
			case ".yml", ".yaml":
				err := yaml.Unmarshal(b, &values)
				if err != nil {
					return nil, fmt.Errorf("invalid YAML on module-variables file %s: %w", pmv, err)
				}
			case ".json":
				err = json.Unmarshal(b, &values)
				if err != nil {
					return nil, fmt.Errorf("invalid JSON on module-variables file %s: %w", pmv, err)
				}
			default:
				return nil, fmt.Errorf("invalid module-variables %s, only supported extensions are yaml/yml/json", pmv)
			}

			for k, v := range values {
				for _, vv := range v {
					mv[fmt.Sprintf("%s.%s", k, vv)] = struct{}{}
				}
			}
		}
	}

	return &writer.Options{
		Interpolate:      viper.GetBool("interpolate"),
		Module:           module,
		ModuleVariables:  mv,
		HCLProviderBlock: viper.GetBool("hcl-provider-block"),
	}, nil
}

func importProvider(ctx context.Context, logger kitlog.Logger, p provider.Provider, tags []tag.Tag) error {
	f := &filter.Filter{
		Include: include,
		Exclude: exclude,
		Targets: targets,
		Tags:    tags,
	}

	var hclW, stateW writer.Writer
	options, err := getWriterOptions()
	if err != nil {
		return err
	}

	if hclOut != nil {
		logger.Log("msg", "initializing HCL writer")
		hclW = hcl.NewWriter(hclOut, p, options)
	}

	if stateOut != nil {
		logger.Log("msg", "initializing TFState writer")
		stateW = state.NewWriter(stateOut, options)
	}

	logger.Log("msg", "importing")

	fmt.Fprintf(logsOut, "Starting Terracognita with version %s\n", Version)
	logger.Log("msg", "starting terracognita", "version", Version)
	err = provider.Import(ctx, p, hclW, stateW, f, logsOut)
	if err != nil {
		return errors.Wrap(err, "could not import from "+p.String())
	}

	return nil
}

// initializeTags returns the list of tags for the flagName, as different
// providers have diferent for them (google names them lables) we need to
// know the actual name of the flag
func initializeTags(flagName string) ([]tag.Tag, error) {
	// Initialize the tags
	tags := make([]tag.Tag, 0, len(viper.GetStringSlice(flagName)))
	for _, t := range viper.GetStringSlice(flagName) {
		tg, err := tag.New(t)
		if err != nil {
			return nil, fmt.Errorf("invalid format for --%s with value %q: %w", flagName, t, err)
		}
		tags = append(tags, tg)
	}
	return tags, nil
}

func init() {
	cobra.OnInitialize(initViper)
	RootCmd.AddCommand(awsCmd)
	RootCmd.AddCommand(googleCmd)
	RootCmd.AddCommand(azurermCmd)
	RootCmd.AddCommand(vsphereCmd)
	RootCmd.AddCommand(versionCmd)

	RootCmd.PersistentFlags().String("hcl", "", "HCL output file or directory. If it's a directory it'll be emptied before importing")
	_ = viper.BindPFlag("hcl", RootCmd.PersistentFlags().Lookup("hcl"))

	RootCmd.PersistentFlags().String("tfstate", "", "TFState output file")
	_ = viper.BindPFlag("tfstate", RootCmd.PersistentFlags().Lookup("tfstate"))

	RootCmd.PersistentFlags().String("module", "", "Generates the output in module format into the directory specified. With this flag (--module) the --hcl is ignored and will be generated inside of the module")
	_ = viper.BindPFlag("module", RootCmd.PersistentFlags().Lookup("module"))

	RootCmd.PersistentFlags().String("module-variables", "", "Path to a file containing the list of attributes to use as variables when building the module. The format is a JSON/YAML, more information on https://github.com/cycloidio/terracognita#modules")
	_ = viper.BindPFlag("module-variables", RootCmd.PersistentFlags().Lookup("module-variables"))

	RootCmd.PersistentFlags().StringSliceVarP(&include, "include", "i", []string{}, "List of resources to import, this names are the ones on TF (ex: aws_instance). If not set then means that all the resources will be imported")
	_ = viper.BindPFlag("include", RootCmd.PersistentFlags().Lookup("include"))

	RootCmd.PersistentFlags().StringSliceVarP(&exclude, "exclude", "e", []string{}, "List of resources to not import, this names are the ones on TF (ex: aws_instance). If not set then means that none the resources will be excluded")
	_ = viper.BindPFlag("exclude", RootCmd.PersistentFlags().Lookup("exclude"))

	RootCmd.PersistentFlags().StringSliceVar(&targets, "target", []string{}, "List of resources to import via ID, those IDs are the ones documented on Terraform that are needed to Import. The format is 'aws_instance.ID'")
	_ = viper.BindPFlag("target", RootCmd.PersistentFlags().Lookup("target"))

	RootCmd.PersistentFlags().BoolP("verbose", "v", false, "Activate the verbose mode")
	_ = viper.BindPFlag("verbose", RootCmd.PersistentFlags().Lookup("verbose"))

	RootCmd.PersistentFlags().BoolP("debug", "d", false, "Activate the debug mode which includes TF logs via TF_LOG=TRACE|DEBUG|INFO|WARN|ERROR configuration https://www.terraform.io/docs/internals/debugging.html")
	_ = viper.BindPFlag("debug", RootCmd.PersistentFlags().Lookup("debug"))

	RootCmd.PersistentFlags().String("log-file", path.Join(xdg.CacheHome, "terracognita", "terracognita.log"), "Write the logs with -v to this destination")
	_ = viper.BindPFlag("log-file", RootCmd.PersistentFlags().Lookup("log-file"))

	RootCmd.PersistentFlags().BoolP("interpolate", "", true, "Activate the interpolation for the HCL and the dependencies building for the State file")
	_ = viper.BindPFlag("interpolate", RootCmd.PersistentFlags().Lookup("interpolate"))

	RootCmd.PersistentFlags().BoolP("hcl-provider-block", "", true, "Generate or not the 'provider {}' block for the imported provider")
	_ = viper.BindPFlag("hcl-provider-block", RootCmd.PersistentFlags().Lookup("hcl-provider-block"))
}

func initViper() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
}
