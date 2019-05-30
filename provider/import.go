package provider

import (
	"context"

	"github.com/cycloidio/terraforming/errcode"
	"github.com/cycloidio/terraforming/filter"
	"github.com/cycloidio/terraforming/writer"
	"github.com/pkg/errors"
)

// Import imports from the Provider p all the resources filtered by f and writes
// the result to the hcl or tfstate if those are not nil
func Import(ctx context.Context, p Provider, hcl, tfstate writer.Writer, f filter.Filter) error {
	var types []string

	if len(f.Include) != 0 {
		types = f.Include
	} else {
		types = p.ResourceTypes()
	}

	for _, t := range types {
		if f.IsExcluded(t) {
			continue
		}
		resources, err := p.Resources(ctx, t, f)
		if err != nil {
			return errors.WithStack(err)
		}

		for _, r := range resources {
			err := r.Read(f)
			// TODO Validate the 2 errors that we do not want to return
			if err != nil {
				if errors.Cause(err) != errcode.ErrResourceNotRead && errors.Cause(err) != errcode.ErrResourceDoNotMatchTag {
					return errors.Wrapf(err, "could not read resource %s: ", r.Type)
				}
				if errors.Cause(err) == errcode.ErrResourceNotRead {
					// As the resource could not be Read, meaning an ID == ""
					// we'll continue to the next resource
					continue
				}
			}

			if hcl != nil {
				err = r.HCL(hcl)
				if err != nil {
					return errors.Wrapf(err, "error while calculating the Config of resource %q", t)
				}
			}

			if tfstate != nil {
				err = r.State(tfstate)
				if err != nil {
					return errors.Wrapf(err, "error while calculating the Satate of resource %q", t)
				}
			}
		}
	}

	if hcl != nil {
		err := hcl.Sync()
		if err != nil {
			return errors.Wrapf(err, "error while Sync Config")
		}
	}

	if tfstate != nil {
		err := tfstate.Sync()
		if err != nil {
			return errors.Wrapf(err, "error while Sync State")
		}
	}

	return nil
}
