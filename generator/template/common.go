package template

import (
	"strings"

	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/util"
	"github.com/vetcher/godecl/types"
)

const (
	PackagePathGoKitEndpoint      = "github.com/go-kit/kit/endpoint"
	PackagePathContext            = "context"
	PackagePathGoKitLog           = "github.com/go-kit/kit/log"
	PackagePathTime               = "time"
	PackagePathGoogleGRPC         = "google.golang.org/grpc"
	PackagePathGoogleGRPCCodes    = "google.golang.org/grpc/codes"
	PackagePathNetContext         = "golang.org/x/net/context"
	PackagePathGoKitTransportGRPC = "github.com/go-kit/kit/transport/grpc"
	PackagePathHttp               = "net/http"
	PackagePathGoKitTransportHTTP = "github.com/go-kit/kit/transport/http"
	PackagePathBytes              = "bytes"
	PackagePathJson               = "encoding/json"
	PackagePathIOUtil             = "io/ioutil"
	PackagePathStrings            = "strings"
	PackagePathUrl                = "net/url"
	PackagePathEmptyProtobuf      = "github.com/golang/protobuf/ptypes/empty"
	PackagePathFmt                = "fmt"
	PackagePathOs                 = "os"
	PackagePathOsSignal           = "os/signal"
	PackagePathSyscall            = "syscall"
	PackagePathErrors             = "errors"
	PackagePathNet                = "net"

	TagMark         = "// @"
	MicrogenMainTag = "microgen"
	ForceTag        = "force"

	Version    = "0.6.0"
	FileHeader = `This file was automatically generated by "microgen ` + Version + `" utility.`
)

const (
	MiddlewareTag        = "middleware"
	LoggingMiddlewareTag = "logging"
	RecoverMiddlewareTag = "recover"
	HttpTag              = "http"
	HttpServerTag        = "http-server"
	HttpClientTag        = "http-client"
	GrpcTag              = "grpc"
	GrpcServerTag        = "grpc-server"
	GrpcClientTag        = "grpc-client"
	MainTag              = "main"
)

type WriteStrategyState int

const (
	FileStrat WriteStrategyState = iota + 1
	AppendStrat
)

type GenerationInfo struct {
	ServiceImportPackageName string
	Iface                    *types.Interface
	ServiceImportPath        string
	Force                    bool
	AbsOutPath               string
	SourceFilePath           string

	ProtobufPackage string
	GRPCRegAddr     string
}

func (info GenerationInfo) Copy() *GenerationInfo {
	return &GenerationInfo{
		Iface: info.Iface,
		Force: info.Force,
		ServiceImportPackageName: info.ServiceImportPackageName,
		ServiceImportPath:        info.ServiceImportPath,
		AbsOutPath:               info.AbsOutPath,
		SourceFilePath:           info.SourceFilePath,

		GRPCRegAddr:     info.GRPCRegAddr,
		ProtobufPackage: info.ProtobufPackage,
	}
}

func structFieldName(field *types.Variable) *Statement {
	return Id(util.ToUpperFirst(field.Name))
}

// Remove from function fields context if it is first in slice
func removeContextIfFirst(fields []types.Variable) []types.Variable {
	if IsContextFirst(fields) {
		return fields[1:]
	}
	return fields
}

func IsContextFirst(fields []types.Variable) bool {
	name := types.TypeName(fields[0].Type)
	return name != nil && len(fields) > 0 &&
		types.TypeImport(fields[0].Type) != nil &&
		types.TypeImport(fields[0].Type).Package == PackagePathContext &&
		*name == "Context"
}

// Remove from function fields error if it is last in slice
func removeErrorIfLast(fields []types.Variable) []types.Variable {
	if IsErrorLast(fields) {
		return fields[:len(fields)-1]
	}
	return fields
}

func IsErrorLast(fields []types.Variable) bool {
	name := types.TypeName(fields[len(fields)-1].Type)
	return name != nil && len(fields) > 0 &&
		types.TypeImport(fields[len(fields)-1].Type) == nil &&
		*name == "error"
}

// Return name of error, if error is last result, else return `err`
func nameOfLastResultError(fn *types.Function) string {
	if IsErrorLast(fn.Results) {
		return fn.Results[len(fn.Results)-1].Name
	}
	return "err"
}

// Renders struct field.
//
//  	Visit *entity.Visit `json:"visit"`
//
func structField(field *types.Variable) *Statement {
	s := structFieldName(field)
	s.Add(fieldType(field.Type, false))
	s.Tag(map[string]string{"json": util.ToSnakeCase(field.Name)})
	if types.IsEllipsis(field.Type) {
		s.Comment("This field was defined with ellipsis (...).")
	}
	return s
}

// Renders func params for definition.
//
//  	visit *entity.Visit, err error
//
func funcDefinitionParams(fields []types.Variable) *Statement {
	c := &Statement{}
	c.ListFunc(func(g *Group) {
		for _, field := range fields {
			g.Id(util.ToLowerFirst(field.Name)).Add(fieldType(field.Type, true))
		}
	})
	return c
}

// Renders field type for given func field.
//
//  	*repository.Visit
//
func fieldType(field types.Type, useEllipsis bool) *Statement {
	c := &Statement{}
	for field != nil {
		switch f := field.(type) {
		case types.TImport:
			if f.Import != nil {
				c.Qual(f.Import.Package, "")
			}
			field = f.Next
		case types.TName:
			c.Id(f.TypeName)
			field = nil
		case types.TArray:
			if f.IsSlice {
				c.Index()
			} else if f.ArrayLen > 0 {
				c.Index(Lit(f.ArrayLen))
			}
			field = f.Next
		case types.TMap:
			return c.Map(fieldType(f.Key, false)).Add(fieldType(f.Value, false))
		case types.TPointer:
			c.Op(strings.Repeat("*", f.NumberOfPointers))
			field = f.Next
		case types.TInterface:
			mhds := interfaceType(f.Interface)
			return c.Interface(mhds...)
		case types.TEllipsis:
			if useEllipsis {
				c.Op("...")
			} else {
				c.Index()
			}
			field = f.Next
		default:
			return c
		}
	}
	return c
}

func interfaceType(p *types.Interface) (code []Code) {
	for _, x := range p.Methods {
		code = append(code, functionDefinition(x))
	}
	return
}

// Renders key/value pairs wrapped in Dict for provided fields.
//
//		Err:    err,
//		Result: result,
//
func dictByVariables(fields []types.Variable) Dict {
	return DictFunc(func(d Dict) {
		for _, field := range fields {
			d[structFieldName(&field)] = Id(util.ToLowerFirst(field.Name))
		}
	})
}

// Render list of function receivers by signature.Result.
//
//		Ans1, ans2, AnS3 -> ans1, ans2, anS3
//
func paramNames(fields []types.Variable) *Statement {
	var list []Code
	for _, field := range fields {
		v := Id(util.ToLowerFirst(field.Name))
		if types.IsEllipsis(field.Type) {
			v.Op("...")
		}
		list = append(list, v)
	}
	return List(list...)
}

// Render full method definition with receiver, method name, args and results.
//
//		func (e *Endpoints) Count(ctx context.Context, text string, symbol string) (count int)
//
func methodDefinition(obj string, signature *types.Function) *Statement {
	return Func().
		Params(Id(util.LastUpperOrFirst(obj)).Op("*").Id(obj)).
		Add(functionDefinition(signature))
}

// Render full method definition with receiver, method name, args and results.
//
//		func Count(ctx context.Context, text string, symbol string) (count int)
//
func functionDefinition(signature *types.Function) *Statement {
	return Id(signature.Name).
		Params(funcDefinitionParams(signature.Args)).
		Params(funcDefinitionParams(signature.Results))
}

// Remove from generating functions that already in existing.
func removeAlreadyExistingFunctions(existing []types.Function, generating *[]*types.Function, nameFormer func(*types.Function) string) {
	x := (*generating)[:0]
	for _, fn := range *generating {
		if f := util.FindFunctionByName(existing, nameFormer(fn)); f == nil {
			x = append(x, fn)
		}
	}
	*generating = x
}
