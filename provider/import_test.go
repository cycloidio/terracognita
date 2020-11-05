package provider_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/cycloidio/terracognita/errcode"
	"github.com/cycloidio/terracognita/filter"
	"github.com/cycloidio/terracognita/mock"
	"github.com/cycloidio/terracognita/provider"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImport(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var (
			ctrl = gomock.NewController(t)
			ctx  = context.Background()

			p                 = mock.NewProvider(ctrl)
			hw                = mock.NewWriter(ctrl)
			sw                = mock.NewWriter(ctrl)
			i                 = make(map[string]string)
			instanceResource1 = mock.NewResource(ctrl)
			instanceResource2 = mock.NewResource(ctrl)
			iamUser1          = mock.NewResource(ctrl)
			iamUser2          = mock.NewResource(ctrl)

			f = &filter.Filter{}
		)

		defer ctrl.Finish()

		p.EXPECT().ResourceTypes().Return([]string{"aws_instance", "aws_iam_user"})

		p.EXPECT().Resources(ctx, "aws_instance", f).Return([]provider.Resource{instanceResource1, instanceResource2}, nil)
		p.EXPECT().Resources(ctx, "aws_iam_user", f).Return([]provider.Resource{iamUser1, iamUser2}, nil)

		instanceResource1.EXPECT().ID().Return("1")
		instanceResource2.EXPECT().ID().Return("2")
		iamUser1.EXPECT().ID().Return("3")
		iamUser2.EXPECT().ID().Return("4")

		instanceResource1.EXPECT().ImportState().Return(nil, nil)
		instanceResource2.EXPECT().ImportState().Return(nil, nil)
		iamUser1.EXPECT().ImportState().Return(nil, nil)
		iamUser2.EXPECT().ImportState().Return(nil, nil)

		instanceResource1.EXPECT().Read(f).Return(nil)
		instanceResource2.EXPECT().Read(f).Return(nil)
		iamUser1.EXPECT().Read(f).Return(nil)
		iamUser2.EXPECT().Read(f).Return(nil)

		instanceResource1.EXPECT().HCL(hw).Return(nil)
		instanceResource2.EXPECT().HCL(hw).Return(nil)
		iamUser1.EXPECT().HCL(hw).Return(nil)
		iamUser2.EXPECT().HCL(hw).Return(nil)

		instanceResource1.EXPECT().State(sw).Return(nil)
		instanceResource2.EXPECT().State(sw).Return(nil)
		iamUser1.EXPECT().State(sw).Return(nil)
		iamUser2.EXPECT().State(sw).Return(nil)

		instanceResource1.EXPECT().InstanceState().Return(nil)
		instanceResource2.EXPECT().InstanceState().Return(nil)
		iamUser1.EXPECT().InstanceState().Return(nil)
		iamUser2.EXPECT().InstanceState().Return(nil)

		hw.EXPECT().Sync().Return(nil)
		hw.EXPECT().Interpolate(i)
		sw.EXPECT().Sync().Return(nil)
		sw.EXPECT().Interpolate(i)

		err := provider.Import(ctx, p, hw, sw, f, ioutil.Discard)
		require.NoError(t, err)
	})
	t.Run("SuccessWithFilterInclude", func(t *testing.T) {
		var (
			ctrl = gomock.NewController(t)
			ctx  = context.Background()

			p                 = mock.NewProvider(ctrl)
			hw                = mock.NewWriter(ctrl)
			sw                = mock.NewWriter(ctrl)
			instanceResource1 = mock.NewResource(ctrl)
			instanceResource2 = mock.NewResource(ctrl)
			i                 = make(map[string]string)

			f = &filter.Filter{
				Include: []string{"aws_instance"},
			}
		)

		defer ctrl.Finish()

		p.EXPECT().HasResourceType("aws_instance").Return(true)
		p.EXPECT().Resources(ctx, "aws_instance", f).Return([]provider.Resource{instanceResource1, instanceResource2}, nil)

		instanceResource1.EXPECT().ID().Return("1")
		instanceResource2.EXPECT().ID().Return("2")

		instanceResource1.EXPECT().ImportState().Return(nil, nil)
		instanceResource2.EXPECT().ImportState().Return(nil, nil)

		instanceResource1.EXPECT().Read(f).Return(nil)
		instanceResource2.EXPECT().Read(f).Return(nil)

		instanceResource1.EXPECT().HCL(hw).Return(nil)
		instanceResource2.EXPECT().HCL(hw).Return(nil)

		instanceResource1.EXPECT().State(sw).Return(nil)
		instanceResource2.EXPECT().State(sw).Return(nil)

		instanceResource1.EXPECT().InstanceState().Return(nil)
		instanceResource2.EXPECT().InstanceState().Return(nil)

		hw.EXPECT().Sync().Return(nil)
		hw.EXPECT().Interpolate(i)
		sw.EXPECT().Sync().Return(nil)
		sw.EXPECT().Interpolate(i)

		err := provider.Import(ctx, p, hw, sw, f, ioutil.Discard)
		require.NoError(t, err)
	})
	t.Run("SuccessWithExclude", func(t *testing.T) {
		var (
			ctrl = gomock.NewController(t)
			ctx  = context.Background()

			p        = mock.NewProvider(ctrl)
			hw       = mock.NewWriter(ctrl)
			sw       = mock.NewWriter(ctrl)
			iamUser1 = mock.NewResource(ctrl)
			iamUser2 = mock.NewResource(ctrl)
			i        = make(map[string]string)

			f = &filter.Filter{
				Exclude: []string{"aws_instance"},
			}
		)

		defer ctrl.Finish()

		p.EXPECT().HasResourceType("aws_instance").Return(true)
		p.EXPECT().ResourceTypes().Return([]string{"aws_instance", "aws_iam_user"})

		p.EXPECT().Resources(ctx, "aws_iam_user", f).Return([]provider.Resource{iamUser1, iamUser2}, nil)

		iamUser1.EXPECT().ID().Return("1")
		iamUser2.EXPECT().ID().Return("2")

		iamUser1.EXPECT().ImportState().Return(nil, nil)
		iamUser2.EXPECT().ImportState().Return(nil, nil)

		iamUser1.EXPECT().Read(f).Return(nil)
		iamUser2.EXPECT().Read(f).Return(nil)

		iamUser1.EXPECT().HCL(hw).Return(nil)
		iamUser2.EXPECT().HCL(hw).Return(nil)

		iamUser1.EXPECT().State(sw).Return(nil)
		iamUser2.EXPECT().State(sw).Return(nil)

		iamUser1.EXPECT().InstanceState().Return(nil)
		iamUser2.EXPECT().InstanceState().Return(nil)

		hw.EXPECT().Sync().Return(nil)
		hw.EXPECT().Interpolate(i)
		sw.EXPECT().Sync().Return(nil)
		sw.EXPECT().Interpolate(i)

		err := provider.Import(ctx, p, hw, sw, f, ioutil.Discard)
		require.NoError(t, err)
	})
	t.Run("SuccessWithErrProviderResourceDoNotMatchTag", func(t *testing.T) {
		var (
			ctrl = gomock.NewController(t)
			ctx  = context.Background()

			p        = mock.NewProvider(ctrl)
			hw       = mock.NewWriter(ctrl)
			sw       = mock.NewWriter(ctrl)
			iamUser1 = mock.NewResource(ctrl)
			iamUser2 = mock.NewResource(ctrl)
			i        = make(map[string]string)

			f = &filter.Filter{
				Exclude: []string{"aws_instance"},
			}
		)

		defer ctrl.Finish()

		p.EXPECT().HasResourceType("aws_instance").Return(true)
		p.EXPECT().ResourceTypes().Return([]string{"aws_instance", "aws_iam_user"})

		p.EXPECT().Resources(ctx, "aws_iam_user", f).Return([]provider.Resource{iamUser1, iamUser2}, nil)

		iamUser1.EXPECT().ID().Return("1")
		iamUser2.EXPECT().ID().Return("2")

		iamUser1.EXPECT().ImportState().Return(nil, nil)
		iamUser2.EXPECT().ImportState().Return(nil, nil)

		iamUser2.EXPECT().InstanceState().Return(nil)

		iamUser1.EXPECT().Read(f).Return(errcode.ErrProviderResourceDoNotMatchTag)
		iamUser2.EXPECT().Read(f).Return(nil)

		iamUser2.EXPECT().HCL(hw).Return(nil)

		iamUser2.EXPECT().State(sw).Return(nil)

		hw.EXPECT().Sync().Return(nil)
		hw.EXPECT().Interpolate(i)
		sw.EXPECT().Sync().Return(nil)
		sw.EXPECT().Interpolate(i)

		err := provider.Import(ctx, p, hw, sw, f, ioutil.Discard)
		require.NoError(t, err)
	})
	t.Run("SuccessWithNoHCLWriter", func(t *testing.T) {
		var (
			ctrl = gomock.NewController(t)
			ctx  = context.Background()

			p        = mock.NewProvider(ctrl)
			sw       = mock.NewWriter(ctrl)
			iamUser1 = mock.NewResource(ctrl)
			iamUser2 = mock.NewResource(ctrl)
			i        = make(map[string]string)

			f = &filter.Filter{
				Exclude: []string{"aws_instance"},
			}
		)

		defer ctrl.Finish()

		p.EXPECT().HasResourceType("aws_instance").Return(true)
		p.EXPECT().ResourceTypes().Return([]string{"aws_instance", "aws_iam_user"})

		p.EXPECT().Resources(ctx, "aws_iam_user", f).Return([]provider.Resource{iamUser1, iamUser2}, nil)

		iamUser1.EXPECT().ID().Return("1")
		iamUser2.EXPECT().ID().Return("2")

		iamUser1.EXPECT().ImportState().Return(nil, nil)
		iamUser2.EXPECT().ImportState().Return(nil, nil)

		iamUser1.EXPECT().Read(f).Return(errcode.ErrProviderResourceDoNotMatchTag)
		iamUser2.EXPECT().Read(f).Return(nil)

		iamUser2.EXPECT().State(sw).Return(nil)
		iamUser2.EXPECT().InstanceState().Return(nil)

		sw.EXPECT().Sync().Return(nil)
		sw.EXPECT().Interpolate(i)

		err := provider.Import(ctx, p, nil, sw, f, ioutil.Discard)
		require.NoError(t, err)
	})
	t.Run("SuccessWithNoTFStateWriter", func(t *testing.T) {
		var (
			ctrl = gomock.NewController(t)
			ctx  = context.Background()

			p        = mock.NewProvider(ctrl)
			hw       = mock.NewWriter(ctrl)
			iamUser1 = mock.NewResource(ctrl)
			iamUser2 = mock.NewResource(ctrl)
			i        = make(map[string]string)

			f = &filter.Filter{
				Exclude: []string{"aws_instance"},
			}
		)

		defer ctrl.Finish()

		p.EXPECT().HasResourceType("aws_instance").Return(true)
		p.EXPECT().ResourceTypes().Return([]string{"aws_instance", "aws_iam_user"})

		p.EXPECT().Resources(ctx, "aws_iam_user", f).Return([]provider.Resource{iamUser1, iamUser2}, nil)

		iamUser1.EXPECT().ID().Return("1")
		iamUser2.EXPECT().ID().Return("2")

		iamUser1.EXPECT().ImportState().Return(nil, nil)
		iamUser2.EXPECT().ImportState().Return(nil, nil)

		iamUser1.EXPECT().Read(f).Return(errcode.ErrProviderResourceDoNotMatchTag)
		iamUser2.EXPECT().Read(f).Return(nil)

		iamUser2.EXPECT().HCL(hw).Return(nil)
		iamUser2.EXPECT().InstanceState().Return(nil)

		hw.EXPECT().Sync().Return(nil)
		hw.EXPECT().Interpolate(i)

		err := provider.Import(ctx, p, hw, nil, f, ioutil.Discard)
		require.NoError(t, err)
	})
	t.Run("ErrorWithErrProviderResourceNotRead", func(t *testing.T) {
		var (
			ctrl = gomock.NewController(t)
			ctx  = context.Background()

			p        = mock.NewProvider(ctrl)
			hw       = mock.NewWriter(ctrl)
			sw       = mock.NewWriter(ctrl)
			iamUser1 = mock.NewResource(ctrl)
			iamUser2 = mock.NewResource(ctrl)
			i        = make(map[string]string)

			f = &filter.Filter{
				Exclude: []string{"aws_instance"},
			}
		)

		defer ctrl.Finish()

		p.EXPECT().HasResourceType("aws_instance").Return(true)
		p.EXPECT().ResourceTypes().Return([]string{"aws_instance", "aws_iam_user"})

		p.EXPECT().Resources(ctx, "aws_iam_user", f).Return([]provider.Resource{iamUser1, iamUser2}, nil)

		iamUser1.EXPECT().ID().Return("1")
		iamUser2.EXPECT().ID().Return("2")

		iamUser1.EXPECT().ImportState().Return(nil, nil)
		iamUser2.EXPECT().ImportState().Return(nil, nil)

		iamUser1.EXPECT().Read(f).Return(errcode.ErrProviderResourceNotRead)
		iamUser2.EXPECT().Read(f).Return(nil)

		iamUser2.EXPECT().HCL(hw).Return(nil)

		iamUser2.EXPECT().State(sw).Return(nil)
		iamUser2.EXPECT().InstanceState().Return(nil)

		hw.EXPECT().Sync().Return(nil)
		hw.EXPECT().Interpolate(i)
		sw.EXPECT().Sync().Return(nil)
		sw.EXPECT().Interpolate(i)

		err := provider.Import(ctx, p, hw, sw, f, ioutil.Discard)
		require.NoError(t, err)
	})
	t.Run("ErrorWithErrProviderResourceAutogenerated", func(t *testing.T) {
		var (
			ctrl = gomock.NewController(t)
			ctx  = context.Background()

			p        = mock.NewProvider(ctrl)
			hw       = mock.NewWriter(ctrl)
			sw       = mock.NewWriter(ctrl)
			iamUser1 = mock.NewResource(ctrl)
			iamUser2 = mock.NewResource(ctrl)
			i        = make(map[string]string)

			f = &filter.Filter{
				Exclude: []string{"aws_instance"},
			}
		)

		defer ctrl.Finish()

		p.EXPECT().HasResourceType("aws_instance").Return(true)
		p.EXPECT().ResourceTypes().Return([]string{"aws_instance", "aws_iam_user"})

		p.EXPECT().Resources(ctx, "aws_iam_user", f).Return([]provider.Resource{iamUser1, iamUser2}, nil)

		iamUser1.EXPECT().ID().Return("1")
		iamUser2.EXPECT().ID().Return("2")

		iamUser1.EXPECT().ImportState().Return(nil, nil)
		iamUser2.EXPECT().ImportState().Return(nil, nil)

		iamUser1.EXPECT().Read(f).Return(errcode.ErrProviderResourceAutogenerated)
		iamUser2.EXPECT().Read(f).Return(nil)

		iamUser2.EXPECT().HCL(hw).Return(nil)

		iamUser2.EXPECT().State(sw).Return(nil)
		iamUser2.EXPECT().InstanceState().Return(nil)

		hw.EXPECT().Sync().Return(nil)
		hw.EXPECT().Interpolate(i)
		sw.EXPECT().Sync().Return(nil)
		sw.EXPECT().Interpolate(i)

		err := provider.Import(ctx, p, hw, sw, f, ioutil.Discard)
		require.NoError(t, err)
	})
	t.Run("ErrorWithIncorrectFilterInclude", func(t *testing.T) {
		var (
			ctrl = gomock.NewController(t)
			ctx  = context.Background()

			p  = mock.NewProvider(ctrl)
			hw = mock.NewWriter(ctrl)
			sw = mock.NewWriter(ctrl)

			f = &filter.Filter{
				Include: []string{"aws_instance", "aws_potato"},
			}
		)

		defer ctrl.Finish()

		p.EXPECT().HasResourceType("aws_instance").Return(true)
		p.EXPECT().HasResourceType("aws_potato").Return(false)

		err := provider.Import(ctx, p, hw, sw, f, ioutil.Discard)
		assert.Equal(t, errcode.ErrProviderResourceNotSupported.Error(), errors.Cause(err).Error())
	})

	t.Run("ErrorWithIncorrectFilterExclude", func(t *testing.T) {
		var (
			ctrl = gomock.NewController(t)
			ctx  = context.Background()

			p  = mock.NewProvider(ctrl)
			hw = mock.NewWriter(ctrl)
			sw = mock.NewWriter(ctrl)

			f = &filter.Filter{
				Exclude: []string{"aws_instance", "aws_potato"},
			}
		)

		defer ctrl.Finish()

		p.EXPECT().ResourceTypes().Return([]string{})
		p.EXPECT().HasResourceType("aws_instance").Return(true)
		p.EXPECT().HasResourceType("aws_potato").Return(false)

		err := provider.Import(ctx, p, hw, sw, f, ioutil.Discard)
		assert.Equal(t, errcode.ErrProviderResourceNotSupported.Error(), errors.Cause(err).Error())
	})
	t.Run("ErrorWithNotErrProviderAPI", func(t *testing.T) {
		var (
			ctrl = gomock.NewController(t)
			ctx  = context.Background()

			p  = mock.NewProvider(ctrl)
			hw = mock.NewWriter(ctrl)
			sw = mock.NewWriter(ctrl)

			f = &filter.Filter{
				Exclude: []string{"aws_instance"},
			}
		)

		defer ctrl.Finish()

		p.EXPECT().HasResourceType("aws_instance").Return(true)
		p.EXPECT().ResourceTypes().Return([]string{"aws_instance", "aws_iam_user"})

		p.EXPECT().Resources(ctx, "aws_iam_user", f).Return(nil, errors.New("should stop the import"))

		err := provider.Import(ctx, p, hw, sw, f, ioutil.Discard)
		assert.Contains(t, err.Error(), "stop the import")
	})
	t.Run("ErrorWithErrProviderAPI", func(t *testing.T) {
		var (
			ctrl = gomock.NewController(t)
			ctx  = context.Background()

			p                 = mock.NewProvider(ctrl)
			hw                = mock.NewWriter(ctrl)
			sw                = mock.NewWriter(ctrl)
			i                 = make(map[string]string)
			instanceResource1 = mock.NewResource(ctrl)
			instanceResource2 = mock.NewResource(ctrl)

			f = &filter.Filter{}
		)

		defer ctrl.Finish()

		p.EXPECT().ResourceTypes().Return([]string{"aws_instance", "aws_iam_user"})

		p.EXPECT().Resources(ctx, "aws_instance", f).Return([]provider.Resource{instanceResource1, instanceResource2}, nil)
		p.EXPECT().Resources(ctx, "aws_iam_user", f).Return(nil, fmt.Errorf("%w: should not stop the import", errcode.ErrProviderAPI))

		instanceResource1.EXPECT().ID().Return("1")
		instanceResource2.EXPECT().ID().Return("2")

		instanceResource1.EXPECT().ImportState().Return(nil, nil)
		instanceResource2.EXPECT().ImportState().Return(nil, nil)

		instanceResource1.EXPECT().Read(f).Return(nil)
		instanceResource2.EXPECT().Read(f).Return(nil)

		instanceResource1.EXPECT().HCL(hw).Return(nil)
		instanceResource2.EXPECT().HCL(hw).Return(nil)

		instanceResource1.EXPECT().State(sw).Return(nil)
		instanceResource2.EXPECT().State(sw).Return(nil)

		instanceResource1.EXPECT().InstanceState().Return(nil)
		instanceResource2.EXPECT().InstanceState().Return(nil)

		hw.EXPECT().Sync().Return(nil)
		hw.EXPECT().Interpolate(i)
		sw.EXPECT().Sync().Return(nil)
		sw.EXPECT().Interpolate(i)

		err := provider.Import(ctx, p, hw, sw, f, ioutil.Discard)
		require.NoError(t, err)
	})
}
