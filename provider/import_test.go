package provider_test

import (
	"context"
	"fmt"
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

			p                = mock.NewProvider(ctrl)
			hw               = mock.NewWriter(ctrl)
			sw               = mock.NewWriter(ctrl)
			instanceResoure1 = mock.NewResource(ctrl)
			instanceResoure2 = mock.NewResource(ctrl)
			iamUser1         = mock.NewResource(ctrl)
			iamUser2         = mock.NewResource(ctrl)

			f = &filter.Filter{}
		)

		defer ctrl.Finish()

		p.EXPECT().ResourceTypes().Return([]string{"aws_instance", "aws_iam_user"})

		p.EXPECT().Resources(ctx, "aws_instance", f).Return([]provider.Resource{instanceResoure1, instanceResoure2}, nil)
		p.EXPECT().Resources(ctx, "aws_iam_user", f).Return([]provider.Resource{iamUser1, iamUser2}, nil)

		instanceResoure1.EXPECT().Read(f).Return(nil)
		instanceResoure2.EXPECT().Read(f).Return(nil)
		iamUser1.EXPECT().Read(f).Return(nil)
		iamUser2.EXPECT().Read(f).Return(nil)

		instanceResoure1.EXPECT().HCL(hw).Return(nil)
		instanceResoure2.EXPECT().HCL(hw).Return(nil)
		iamUser1.EXPECT().HCL(hw).Return(nil)
		iamUser2.EXPECT().HCL(hw).Return(nil)

		instanceResoure1.EXPECT().State(sw).Return(nil)
		instanceResoure2.EXPECT().State(sw).Return(nil)
		iamUser1.EXPECT().State(sw).Return(nil)
		iamUser2.EXPECT().State(sw).Return(nil)

		hw.EXPECT().Sync().Return(nil)
		sw.EXPECT().Sync().Return(nil)

		err := provider.Import(ctx, p, hw, sw, f)
		require.NoError(t, err)
	})
	t.Run("SuccessWithFilterInclude", func(t *testing.T) {
		var (
			ctrl = gomock.NewController(t)
			ctx  = context.Background()

			p                = mock.NewProvider(ctrl)
			hw               = mock.NewWriter(ctrl)
			sw               = mock.NewWriter(ctrl)
			instanceResoure1 = mock.NewResource(ctrl)
			instanceResoure2 = mock.NewResource(ctrl)

			f = &filter.Filter{
				Include: []string{"aws_instance"},
			}
		)

		defer ctrl.Finish()

		p.EXPECT().HasResourceType("aws_instance").Return(true)
		p.EXPECT().Resources(ctx, "aws_instance", f).Return([]provider.Resource{instanceResoure1, instanceResoure2}, nil)

		instanceResoure1.EXPECT().Read(f).Return(nil)
		instanceResoure2.EXPECT().Read(f).Return(nil)

		instanceResoure1.EXPECT().HCL(hw).Return(nil)
		instanceResoure2.EXPECT().HCL(hw).Return(nil)

		instanceResoure1.EXPECT().State(sw).Return(nil)
		instanceResoure2.EXPECT().State(sw).Return(nil)

		hw.EXPECT().Sync().Return(nil)
		sw.EXPECT().Sync().Return(nil)

		err := provider.Import(ctx, p, hw, sw, f)
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

			f = &filter.Filter{
				Exclude: []string{"aws_instance"},
			}
		)

		defer ctrl.Finish()

		p.EXPECT().ResourceTypes().Return([]string{"aws_instance", "aws_iam_user"})

		p.EXPECT().Resources(ctx, "aws_iam_user", f).Return([]provider.Resource{iamUser1, iamUser2}, nil)

		iamUser1.EXPECT().Read(f).Return(nil)
		iamUser2.EXPECT().Read(f).Return(nil)

		iamUser1.EXPECT().HCL(hw).Return(nil)
		iamUser2.EXPECT().HCL(hw).Return(nil)

		iamUser1.EXPECT().State(sw).Return(nil)
		iamUser2.EXPECT().State(sw).Return(nil)

		hw.EXPECT().Sync().Return(nil)
		sw.EXPECT().Sync().Return(nil)

		err := provider.Import(ctx, p, hw, sw, f)
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

			f = &filter.Filter{
				Exclude: []string{"aws_instance"},
			}
		)

		defer ctrl.Finish()

		p.EXPECT().ResourceTypes().Return([]string{"aws_instance", "aws_iam_user"})

		p.EXPECT().Resources(ctx, "aws_iam_user", f).Return([]provider.Resource{iamUser1, iamUser2}, nil)

		iamUser1.EXPECT().Read(f).Return(errcode.ErrProviderResourceDoNotMatchTag)
		iamUser2.EXPECT().Read(f).Return(nil)

		iamUser1.EXPECT().HCL(hw).Return(nil)
		iamUser2.EXPECT().HCL(hw).Return(nil)

		iamUser1.EXPECT().State(sw).Return(nil)
		iamUser2.EXPECT().State(sw).Return(nil)

		hw.EXPECT().Sync().Return(nil)
		sw.EXPECT().Sync().Return(nil)

		err := provider.Import(ctx, p, hw, sw, f)
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

			f = &filter.Filter{
				Exclude: []string{"aws_instance"},
			}
		)

		defer ctrl.Finish()

		p.EXPECT().ResourceTypes().Return([]string{"aws_instance", "aws_iam_user"})

		p.EXPECT().Resources(ctx, "aws_iam_user", f).Return([]provider.Resource{iamUser1, iamUser2}, nil)

		iamUser1.EXPECT().Read(f).Return(errcode.ErrProviderResourceDoNotMatchTag)
		iamUser2.EXPECT().Read(f).Return(nil)

		iamUser1.EXPECT().State(sw).Return(nil)
		iamUser2.EXPECT().State(sw).Return(nil)

		sw.EXPECT().Sync().Return(nil)

		err := provider.Import(ctx, p, nil, sw, f)
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

			f = &filter.Filter{
				Exclude: []string{"aws_instance"},
			}
		)

		defer ctrl.Finish()

		p.EXPECT().ResourceTypes().Return([]string{"aws_instance", "aws_iam_user"})

		p.EXPECT().Resources(ctx, "aws_iam_user", f).Return([]provider.Resource{iamUser1, iamUser2}, nil)

		iamUser1.EXPECT().Read(f).Return(errcode.ErrProviderResourceDoNotMatchTag)
		iamUser2.EXPECT().Read(f).Return(nil)

		iamUser1.EXPECT().HCL(hw).Return(nil)
		iamUser2.EXPECT().HCL(hw).Return(nil)

		hw.EXPECT().Sync().Return(nil)

		err := provider.Import(ctx, p, hw, nil, f)
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

			f = &filter.Filter{
				Exclude: []string{"aws_instance"},
			}
		)

		defer ctrl.Finish()

		p.EXPECT().ResourceTypes().Return([]string{"aws_instance", "aws_iam_user"})

		p.EXPECT().Resources(ctx, "aws_iam_user", f).Return([]provider.Resource{iamUser1, iamUser2}, nil)

		iamUser1.EXPECT().Read(f).Return(errcode.ErrProviderResourceNotRead)
		iamUser2.EXPECT().Read(f).Return(nil)

		iamUser2.EXPECT().HCL(hw).Return(nil)

		iamUser2.EXPECT().State(sw).Return(nil)

		hw.EXPECT().Sync().Return(nil)
		sw.EXPECT().Sync().Return(nil)

		err := provider.Import(ctx, p, hw, sw, f)
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

			f = &filter.Filter{
				Exclude: []string{"aws_instance"},
			}
		)

		defer ctrl.Finish()

		p.EXPECT().ResourceTypes().Return([]string{"aws_instance", "aws_iam_user"})

		p.EXPECT().Resources(ctx, "aws_iam_user", f).Return([]provider.Resource{iamUser1, iamUser2}, nil)

		iamUser1.EXPECT().Read(f).Return(errcode.ErrProviderResourceAutogenerated)
		iamUser2.EXPECT().Read(f).Return(nil)

		iamUser2.EXPECT().HCL(hw).Return(nil)

		iamUser2.EXPECT().State(sw).Return(nil)

		hw.EXPECT().Sync().Return(nil)
		sw.EXPECT().Sync().Return(nil)

		err := provider.Import(ctx, p, hw, sw, f)
		require.NoError(t, err)
	})
	t.Run("ErrorWithIncorrectFilter", func(t *testing.T) {
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

		err := provider.Import(ctx, p, hw, sw, f)
		fmt.Println(err)
		assert.Equal(t, errcode.ErrProviderResourceNotSupported.Error(), errors.Cause(err).Error())
	})
}
