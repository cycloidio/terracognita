package provider

import (
	"context"
	"fmt"
	"io"

	kitlog "github.com/go-kit/kit/log"

	"github.com/cycloidio/terracognita/errcode"
	"github.com/cycloidio/terracognita/filter"
	"github.com/cycloidio/terracognita/interpolator"
	"github.com/cycloidio/terracognita/log"
	"github.com/cycloidio/terracognita/util"
	"github.com/cycloidio/terracognita/writer"
	"github.com/pkg/errors"
)

// Import imports from the Provider p all the resources filtered by f and writes
// the result to the hcl or tfstate if those are not nil
func Import(ctx context.Context, p Provider, hcl, tfstate writer.Writer, f *filter.Filter, out io.Writer) error {
	logger := log.Get()
	logger = kitlog.With(logger, "func", "provider.Import")

	if err := f.Validate(); err != nil {
		return err
	}

	var (
		err          error
		types        []string
		typesWithIDs map[string][]string
	)

	if len(f.Targets) != 0 {
		typesWithIDs = f.TargetsTypesWithIDs()
		for k := range typesWithIDs {
			if !p.HasResourceType(k) {
				return errors.Wrapf(errcode.ErrProviderResourceNotSupported, "type %s on Target filter", k)
			}
			types = append(types, k)
		}
	} else {
		// Validate if the Include filter is right
		if len(f.Include) != 0 {
			for _, i := range f.Include {
				if !p.HasResourceType(i) {
					return errors.Wrapf(errcode.ErrProviderResourceNotSupported, "type %s on Include filter", i)
				}
			}
			types = f.Include
		} else {
			types = p.ResourceTypes()
		}

		// Validate if the Exclude filter is right
		if len(f.Exclude) != 0 {
			for _, e := range f.Exclude {
				if !p.HasResourceType(e) {
					return errors.Wrapf(errcode.ErrProviderResourceNotSupported, "type %s on Exclude filter", e)
				}
			}
		}
	}

	fmt.Fprintf(out, "Importing with filters: %s", f)
	logger.Log("filters", f.String())

	interpolation := interpolator.New(p.String())

	for _, t := range types {
		logger := kitlog.With(logger, "resource", t)

		if f.IsExcluded(t) {
			logger.Log("msg", "excluded")
			continue
		}

		logger.Log("msg", "fetching the list of resources")

		var resources []Resource

		if typesWithIDs != nil {
			for _, ID := range typesWithIDs[t] {
				resources = append(resources, NewResource(ID, t, p))
			}
		} else {
			resources, err = p.Resources(ctx, t, f)
			if err != nil {
				// we filter the error: if it's an error provider side, we continue
				// the import but we print the error.
				if errors.Is(err, errcode.ErrProviderAPI) {
					logger.Log("msg", fmt.Sprintf("unable to import resource %s: %s\n", t, err.Error()))
				} else {
					return errors.WithStack(err)
				}
			}
		}

		resourceLen := len(resources)
		for i, re := range resources {
			logger := kitlog.With(logger, "id", re.ID(), "total", resourceLen, "current", i+1)
			fmt.Fprintf(out, "\rImporting %s [%d/%d]", t, i+1, resourceLen)

			logger.Log("msg", "reading from TF")
			res, err := re.ImportState()
			if err != nil {
				return err
			}

			// If the InstanceState is nil after the ImportState it
			// means that nothing was imported (potentially is not even Importable)
			// so we have to skip the resource
			if re.InstanceState() == nil {
				continue
			}

			// In case there is more than one State to import
			// we create a new slice with those elements and iterate
			// over it
			for _, r := range append([]Resource{re}, res...) {
				err = util.RetryDefault(func() error { return r.Read(f) })
				if err != nil {
					// Errors are ignored. If a resource is invalid we assume it can be skipped, it can be related to inconsistencies in deployed resources.
					// So instead of failing and stopping execution we ignore them and continue (we log them if -v is specified)

					logger.Log("error", err)

					continue
				}

				if hcl != nil {
					logger.Log("msg", "calculating HCL")
					err = r.HCL(hcl)
					if err != nil {
						return errors.Wrapf(err, "error while calculating the Config of resource %q", t)
					}
				}

				if tfstate != nil {
					logger.Log("msg", "calculating TFState")
					err = r.State(tfstate)
					if err != nil {
						return errors.Wrapf(err, "error while calculating the satate of resource %q", t)
					}
				}
				state := r.InstanceState()

				if state != nil {
					attributes, err := re.AttributesReference()
					if err != nil {
						return errors.Wrapf(err, "unable to fetch attributes of resource")
					}
					attrs := make(map[string]string)
					for _, attribute := range attributes {
						value, ok := state.Attributes[attribute]
						if !ok || len(value) == 0 {
							continue
						}
						attrs[attribute] = value
					}
					interpolation.AddResourceAttributes(fmt.Sprintf("%s.%s", r.Type(), r.Name()), attrs)
				}
			}
		}
		if resourceLen > 0 {
			fmt.Fprintf(out, "\rImporting %s [%d/%d] Done!\n", t, resourceLen, resourceLen)
		}
		logger.Log("msg", "importing done")
	}

	if hcl != nil {
		hcl.Interpolate(interpolation)
		fmt.Fprintf(out, "\rWriting HCL ...")
		logger.Log("msg", "writing the HCL")

		err = hcl.Sync()
		if err != nil {
			return errors.Wrapf(err, "error while Sync Config")
		}

		fmt.Fprintf(out, "\rWriting HCL Done!\n")
		logger.Log("msg", "writing the HCL done")
	}

	if tfstate != nil {
		tfstate.Interpolate(interpolation)
		fmt.Fprintf(out, "\rWriting TFState ...")
		logger.Log("msg", "writing the TFState")

		err := tfstate.Sync()
		if err != nil {
			return errors.Wrapf(err, "error while Sync State")
		}

		fmt.Fprintf(out, "\rWriting TFState Done!\n")
		logger.Log("msg", "writing the TFState done")
	}

	return nil
}
