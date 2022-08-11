package parser

import (
	"go/ast"

	"encr.dev/parser/est"
	"encr.dev/parser/internal/locations"
	"encr.dev/parser/internal/walker"
	schema "encr.dev/proto/encore/parser/schema/v1"
)

const configImportPath = "encore.dev/config"

func init() {
	registerResource(
		est.ConfigResource,
		"config",
		"https://encore.dev/docs/develop/config",
		"config",
		configImportPath,
	)

	registerResourceCreationParser(
		est.ConfigResource,
		"Load", 1,
		(*parser).parseConfigLoad,
		locations.AllowedIn(locations.Variable).ButNotIn(locations.Function),
	)

	registerTypeResolver(configImportPath, (*parser).resolveConfigTypes)
}

func (p *parser) parseConfigLoad(file *est.File, _ *walker.Cursor, ident *ast.Ident, callExpr *ast.CallExpr) est.Resource {
	// Resolve the named struct used for the config type
	configType := p.resolveParameter("config type", file.Pkg, file, getTypeArguments(callExpr.Fun)[0])
	if configType == nil {
		return nil
	}

	// Check we have a service
	svc := file.Pkg.Service
	if svc == nil {
		p.err(
			callExpr.Pos(),
			"A call to config.Load[T]() can only be made from within a service.",
		)
		return nil
	}
	if svc.Root != file.Pkg {
		p.err(
			callExpr.Pos(),
			"A call to config.Load[T]() can only be made from the top level package of a service.",
		)
		return nil
	}

	estNode := &est.Config{
		Svc:          file.Pkg.Service,
		DeclFile:     file,
		IdentAST:     ident,
		ConfigStruct: configType,
	}
	svc.ConfigLoads = append(svc.ConfigLoads, estNode)

	return estNode
}

func (p *parser) resolveConfigTypes(ident *ast.Ident, typeParameters typeParameterLookup) *schema.Type {
	switch ident.Name {
	case "Value":
		return &schema.Type{
			Typ: &schema.Type_Config{
				Config: &schema.ConfigValue{
					Elem: nil, // we set this to nil here, and expect the type resolver to set it later
				},
			},
		}

	// Built-in helper types we expose in the config package
	case "Bool":
		return createBuiltinConfigWrapper(schema.Builtin_BOOL)
	case "Int8":
		return createBuiltinConfigWrapper(schema.Builtin_INT8)
	case "Int16":
		return createBuiltinConfigWrapper(schema.Builtin_INT16)
	case "Int32":
		return createBuiltinConfigWrapper(schema.Builtin_INT32)
	case "Int64":
		return createBuiltinConfigWrapper(schema.Builtin_INT64)
	case "UInt8":
		return createBuiltinConfigWrapper(schema.Builtin_UINT8)
	case "UInt16":
		return createBuiltinConfigWrapper(schema.Builtin_UINT16)
	case "UInt32":
		return createBuiltinConfigWrapper(schema.Builtin_UINT32)
	case "UInt64":
		return createBuiltinConfigWrapper(schema.Builtin_UINT64)
	case "Float32":
		return createBuiltinConfigWrapper(schema.Builtin_FLOAT32)
	case "Float64":
		return createBuiltinConfigWrapper(schema.Builtin_FLOAT64)
	case "String":
		return createBuiltinConfigWrapper(schema.Builtin_STRING)
	case "Time":
		return createBuiltinConfigWrapper(schema.Builtin_TIME)
	case "UUID":
		return createBuiltinConfigWrapper(schema.Builtin_UUID)
	case "Int":
		return createBuiltinConfigWrapper(schema.Builtin_INT)
	case "UInt":
		return createBuiltinConfigWrapper(schema.Builtin_UINT)
	default:
		p.errf(ident.Pos(), "config.%s is not type which can be used within data structures", ident.Name)
		return nil
	}
}

func createBuiltinConfigWrapper(builtIn schema.Builtin) *schema.Type {
	return &schema.Type{
		Typ: &schema.Type_Config{
			Config: &schema.ConfigValue{
				Elem: &schema.Type{
					Typ: &schema.Type_Builtin{
						Builtin: builtIn,
					},
				},
			},
		},
	}
}
