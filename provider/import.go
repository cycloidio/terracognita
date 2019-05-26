package provider

import (
	"context"

	"github.com/cycloidio/terraforming/errcode"
	"github.com/cycloidio/terraforming/filter"
	"github.com/cycloidio/terraforming/writer"
	"github.com/pkg/errors"
)

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
			if err != nil && errors.Cause(err) != errcode.ErrResourceNotRead && errors.Cause(err) != errcode.ErrResourceDoNotMatchTag {
				return errors.Wrapf(err, "could not read resource %s: ", r.Type)
			}

			if tfstate != nil {
				r.State(tfstate)
			}
			if hcl != nil {
				r.HCL(hcl)
			}
		}
	}

	if hcl != nil {
		err := hcl.Sync()
		if err != nil {
			return err
		}
	}

	if tfstate != nil {
		err := tfstate.Sync()
		if err != nil {
			return err
		}
	}

	return nil
}
