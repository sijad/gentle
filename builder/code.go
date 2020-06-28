package builder

import (
	"fmt"
	"go/types"

	"github.com/dave/jennifer/jen"
	"github.com/jensneuse/graphql-go-tools/pkg/introspection"
)

func (g *gqlBuilder) Code() string {
	gen := jen.NewFile("generated")

	// TODO changing maps key to int might improve performace
	// gen.Func().
	// 	Id("hashId").
	// 	Params(jen.Id("s").String()).
	// 	Block(
	// 		jen.Id("h").Op(":=").Qual("hash/fnv", "New32a").Call(),
	// 		jen.Id("h").Dot("Write").Call(jen.Index().Byte().Call(jen.Id("s"))),
	// 		jen.Return(jen.Id("h").Dot("Sum32").Call()),
	// 	)

	gen.Type().Id("InputArgs").Map(jen.String()).Interface()
	gen.Type().Id("DependencyInjection").Map(jen.String()).Interface()

	resolverMapParams := []jen.Code{
		jen.Id("root").Interface(),
		jen.Id("args").Op("InputArgs"),
		jen.Id("di").Op("DependencyInjection"),
	}
	for _, ftyp := range g.FullTypes() {
		switch ftyp.Kind {
		case introspection.OBJECT:
			gen.Var().
				Id(ftyp.Name + "Map").
				Op("=").
				Map(jen.String()).
				Func().
				Params(resolverMapParams...).
				Parens(jen.List(jen.Interface(), jen.Error())).
				Values(jen.DictFunc(func(d jen.Dict) {
					rootId := jen.Id("root").Assert(jen.Op("*").Qual(ftyp.PkgPath, ftyp.Name))
					for _, field := range ftyp.Fields {
						d[jen.Lit(field.Name)] = jen.Func().
							Params(resolverMapParams...).
							Parens(jen.List(jen.Interface(), jen.Error())).
							BlockFunc(func(g *jen.Group) {
								if field.IsMethod {
									g.Return(rootId.Clone().Dot(field.Name).Call(jen.ListFunc(func(g *jen.Group) {
										for _, param := range field.Params {
											name := param.Name()
											typ := param.Type()
											if name == "args" {
												var argsStruct *types.Struct
												switch xx := typ.(type) {
												case *types.Named:
													argsStruct = xx.Underlying().(*types.Struct)
												case *types.Struct:
													argsStruct = xx
												default:
													panic("args can only be struct")
												}
												typPkg, typName := typePath(typ)
												var s *jen.Statement
												if typPkg != "" {
													s = g.Qual(typPkg, typName)
												} else {
													s = g.Id(typName)
												}
												s.Values(jen.DictFunc(func(d jen.Dict) {
													for i := 0; i < argsStruct.NumFields(); i++ {
														argField := argsStruct.Field(i)

														if !argField.Exported() {
															continue
														}

														argName := argField.Name()
														d[jen.Id(argName)] = assertType("args", argName, argField.Type())
													}
												}))
												continue
											}
											g.Add(assertType("di", typ.String(), typ))
										}
									})))
								} else {
									g.Return(jen.Nil(), jen.Nil())
								}
							})
					}
				}))
		}
	}
	return fmt.Sprintf("%#v", gen)
}

func assertType(id, index string, typ types.Type) jen.Code {
	typPkg, typName := typePath(typ)
	return jen.Id(id).Index(jen.Lit(index)).Assert(jen.Do(func(s *jen.Statement) {
		if typPkg != "" {
			s.Qual(typPkg, typName)
		} else {
			s.Id(typName)
		}
	}))
}

func typePath(typ types.Type) (string, string) {
	switch x := typ.(type) {
	case *types.Basic:
		return "", x.Name()
	case *types.Named:
		obj := x.Obj()
		return obj.Pkg().Path(), obj.Name()
	default:
		return "", x.String()
	}
}

func typeCode(typ types.Type) jen.Code {
	return nil
}
