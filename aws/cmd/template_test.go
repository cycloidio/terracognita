package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateName(t *testing.T) {
	tests := []struct {
		name string
		tmp  Function
		opt  string
	}{
		{
			name: "Basic",
			tmp: Function{
				Entity: "Entity",
			},
			opt: "GetEntity",
		},
		{
			name: "FnName",
			tmp: Function{
				FnName: "FnEntity",
			},
			opt: "FnEntity",
		},
		{
			name: "FilterByOwner",
			tmp: Function{
				Entity:        "Entity",
				FilterByOwner: "not-relevant",
			},
			opt: "GetOwnEntity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.opt, tt.tmp.Name())
		})
	}
}

func TestTemplateOutput(t *testing.T) {
	tests := []struct {
		name string
		tmp  Function
		opt  string
	}{
		{
			name: "Basic",
			tmp: Function{
				Service: "Service",
				Entity:  "Entity",
			},
			opt: "[]*Service.Entity",
		},
		{
			name: "FnOutput",
			tmp: Function{
				Service:  "Service",
				FnOutput: "FnOutput",
			},
			opt: "[]*FnOutput",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.opt, tt.tmp.Output())
		})
	}
}

func TestTemplateInput(t *testing.T) {
	tests := []struct {
		name string
		tmp  Function
		opt  string
	}{
		{
			name: "Basic",
			tmp: Function{
				Service: "Service",
				Entity:  "Entity",
				Prefix:  "Prefix",
			},
			opt: "Service.PrefixEntityInput",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.opt, tt.tmp.Input())
		})
	}
}

func TestTemplateSignature(t *testing.T) {
	tests := []struct {
		name string
		tmp  Function
		opt  string
	}{
		{
			name: "Basic",
			tmp: Function{
				Service: "Service",
				Entity:  "Entities",
				Prefix:  "Prefix",
			},
			opt: "GetEntities (ctx context.Context, input *Service.PrefixEntitiesInput) ([]*Service.Entity, error)",
		},
		{
			name: "FnSignature",
			tmp: Function{
				FnSignature: "SomeSignature",
			},
			opt: "SomeSignature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.opt, tt.tmp.Signature())
		})
	}
}

func TestTemplateExecute(t *testing.T) {
	tests := []struct {
		name string
		tmp  Function
		opt  string
	}{
		{
			name: "Basic",
			tmp: Function{
				FnSignature: "Signature",
				Service:     "Service",
				Entity:      "Entities",
				Prefix:      "Prefix",
			},
			opt: `
			func (c *connector) Signature {
				if c.svc.Service == nil {
					c.svc.Service = Service.New(c.svc.session)
				}

				opt := make([]*Service.Entity, 0)

				hasNextToken := true
				for hasNextToken {
					o, err := c.svc.Service.PrefixEntitiesWithContext(ctx, input)
					if err != nil {
						return nil, err
					}
					if input == nil {
						input = &Service.PrefixEntitiesInput{}
					}
					input.NextToken = o.NextToken
					hasNextToken = o.NextToken != nil

					opt = append(opt, o.Entities...)
				}

				return opt, nil
			}`,
		},
		{
			name: "FilterByOwner",
			tmp: Function{
				FilterByOwner: "OwnerField",
				FnSignature:   "Signature",
				Service:       "Service",
				Entity:        "Entities",
				Prefix:        "Prefix",
			},
			opt: `
			func (c *connector) Signature {
				if input == nil {
					input = &Service.PrefixEntitiesInput{}
				}
				input.OwnerField = append(input.OwnerField, c.accountID)

				if c.svc.Service == nil {
					c.svc.Service = Service.New(c.svc.session)
				}

				opt := make([]*Service.Entity, 0)

				hasNextToken := true
				for hasNextToken {
					o, err := c.svc.Service.PrefixEntitiesWithContext(ctx, input)
					if err != nil {
						return nil, err
					}
					if input == nil {
						input = &Service.PrefixEntitiesInput{}
					}
					input.NextToken = o.NextToken
					hasNextToken = o.NextToken != nil

					opt = append(opt, o.Entities...)
				}

				return opt, nil
			}`,
		},
		{
			name: "NoGenerateFn",
			tmp: Function{
				NoGenerateFn: true,
			},
			opt: ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buff := bytes.Buffer{}
			err := tt.tmp.Execute(&buff)
			require.NoError(t, err)
			ttopt := strings.Join(strings.Fields(tt.opt), " ")
			buffs := strings.Join(strings.Fields(buff.String()), " ")
			assert.Equal(t, ttopt, buffs)
		})
	}
}
