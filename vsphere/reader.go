package vsphere

import (
	"context"

	"net/url"

	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/session/cache"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/soap"
)

type reader struct {
	*find.Finder
	*view.Manager
}

type optional struct {
	Username string
	Password string
	Insecure bool
}

// newVSphereReader returns new instance of reader.
func newVSphereReader(ctx context.Context, soapURL string, o optional) (*reader, error) {
	u, err := soap.ParseURL(soapURL)
	if err != nil {
		return nil, err
	}

	// Override username if provided
	if o.Username != "" {
		var password string
		var ok bool

		if u.User != nil {
			password, ok = u.User.Password()
		}

		if ok {
			u.User = url.UserPassword(o.Username, password)
		} else {
			u.User = url.User(o.Username)
		}
	}

	// Override password if provided
	if o.Password != "" {
		var username string

		if u.User != nil {
			username = u.User.Username()
		}

		u.User = url.UserPassword(username, o.Password)
	}

	s := &cache.Session{
		URL:      u,
		Insecure: o.Insecure,
	}

	c := new(vim25.Client)
	err = s.Login(ctx, c, nil)
	if err != nil {
		return nil, err
	}

	return &reader{
		Finder: find.NewFinder(c),
	}, nil
}
