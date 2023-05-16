package toproto

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tfplugin6"
)

func Schema(in *tfprotov6.Schema) (*tfplugin6.Schema, error) {
	var resp tfplugin6.Schema
	resp.Version = in.Version
	if in.Block != nil {
		block, err := Schema_Block(in.Block)
		if err != nil {
			return &resp, fmt.Errorf("error marshalling block: %w", err)
		}
		resp.Block = block
	}
	return &resp, nil
}

func Schema_Block(in *tfprotov6.SchemaBlock) (*tfplugin6.Schema_Block, error) {
	resp := &tfplugin6.Schema_Block{
		Version:         in.Version,
		Description:     in.Description,
		DescriptionKind: StringKind(in.DescriptionKind),
		Deprecated:      in.Deprecated,
	}
	attrs, err := Schema_Attributes(in.Attributes)
	if err != nil {
		return resp, err
	}
	resp.Attributes = attrs
	blocks, err := Schema_NestedBlocks(in.BlockTypes)
	if err != nil {
		return resp, err
	}
	resp.BlockTypes = blocks
	return resp, nil
}

func Schema_Attribute(in *tfprotov6.SchemaAttribute) (*tfplugin6.Schema_Attribute, error) {
	resp := &tfplugin6.Schema_Attribute{
		Name:            in.Name,
		Description:     in.Description,
		Required:        in.Required,
		Optional:        in.Optional,
		Computed:        in.Computed,
		Sensitive:       in.Sensitive,
		DescriptionKind: StringKind(in.DescriptionKind),
		Deprecated:      in.Deprecated,
	}
	if in.Type != nil {
		t, err := CtyType(in.Type)
		if err != nil {
			return resp, fmt.Errorf("error marshaling type to JSON: %w", err)
		}
		resp.Type = t
	}
	if in.NestedType != nil {
		nb, err := Schema_Object(in.NestedType)
		if err != nil {
			return resp, err
		}
		resp.NestedType = nb
	}
	return resp, nil
}

func Schema_Attributes(in []*tfprotov6.SchemaAttribute) ([]*tfplugin6.Schema_Attribute, error) {
	resp := make([]*tfplugin6.Schema_Attribute, 0, len(in))
	for _, a := range in {
		if a == nil {
			resp = append(resp, nil)
			continue
		}
		attr, err := Schema_Attribute(a)
		if err != nil {
			return nil, err
		}
		resp = append(resp, attr)
	}
	return resp, nil
}

func Schema_NestedBlock(in *tfprotov6.SchemaNestedBlock) (*tfplugin6.Schema_NestedBlock, error) {
	resp := &tfplugin6.Schema_NestedBlock{
		TypeName: in.TypeName,
		Nesting:  Schema_NestedBlock_NestingMode(in.Nesting),
		MinItems: in.MinItems,
		MaxItems: in.MaxItems,
	}
	if in.Block != nil {
		block, err := Schema_Block(in.Block)
		if err != nil {
			return resp, fmt.Errorf("error marshaling nested block: %w", err)
		}
		resp.Block = block
	}
	return resp, nil
}

func Schema_NestedBlocks(in []*tfprotov6.SchemaNestedBlock) ([]*tfplugin6.Schema_NestedBlock, error) {
	resp := make([]*tfplugin6.Schema_NestedBlock, 0, len(in))
	for _, b := range in {
		if b == nil {
			resp = append(resp, nil)
			continue
		}
		block, err := Schema_NestedBlock(b)
		if err != nil {
			return nil, err
		}
		resp = append(resp, block)
	}
	return resp, nil
}

func Schema_NestedBlock_NestingMode(in tfprotov6.SchemaNestedBlockNestingMode) tfplugin6.Schema_NestedBlock_NestingMode {
	return tfplugin6.Schema_NestedBlock_NestingMode(in)
}

func Schema_Object_NestingMode(in tfprotov6.SchemaObjectNestingMode) tfplugin6.Schema_Object_NestingMode {
	return tfplugin6.Schema_Object_NestingMode(in)
}

func Schema_Object(in *tfprotov6.SchemaObject) (*tfplugin6.Schema_Object, error) {
	resp := &tfplugin6.Schema_Object{
		Nesting:  Schema_Object_NestingMode(in.Nesting),
		MinItems: in.MinItems,
		MaxItems: in.MaxItems,
	}
	attrs, err := Schema_Attributes(in.Attributes)
	if err != nil {
		return nil, err
	}
	resp.Attributes = attrs

	return resp, nil
}

// we have to say this next thing to get golint to stop yelling at us about the
// underscores in the function names. We want the function names to match
// actually-generated code, so it feels like fair play. It's just a shame we
// lose golint for the entire file.
//
// This file is not actually generated. You can edit it. Ignore this next line.
// Code generated by hand ignore this next bit DO NOT EDIT.
