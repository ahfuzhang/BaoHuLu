package codegen

import "github.com/ahfuzhang/BaoHuLu/internal/schema"

type Options struct {
	Source      string
	TemplateDir string
	OutputDir   string
	Version     string
}

type RenderContext struct {
	Schema   *schema.Schema
	File     *schema.File
	Message  *schema.Message
	Enum     *schema.Enum
	Service  *schema.Service
	RPC      *schema.RPC
	TypeRef  *schema.TypeRef
	Analysis *schema.Analysis
	Options  Options
	Args     []any
}

func WithRPC(context RenderContext, method *schema.RPC) RenderContext {
	context.Message, context.Enum, context.TypeRef = nil, nil, nil
	context.RPC = method
	if method != nil {
		context.Service = method.Service
	}
	return context
}

func WithTypeRef(context RenderContext, reference *schema.TypeRef) RenderContext {
	context.TypeRef = reference
	return context
}
