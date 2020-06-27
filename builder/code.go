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

	resolverMapParams := []jen.Code{
		jen.Id("root").Interface(),
		jen.Id("args").Map(jen.String()).Interface(),
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
					rootId := jen.Id("root").Dot("").Parens(jen.Op("*").Qual(ftyp.PkgPath, ftyp.Name))
					for _, field := range ftyp.Fields {
						d[jen.Lit(field.Name)] = jen.Func().
							Params(resolverMapParams...).
							Parens(jen.List(jen.Interface(), jen.Error())).
							BlockFunc(func(g *jen.Group) {
								if field.IsMethod {
									var args jen.Code
									if len(field.Args) > 0 {

									}
									g.Return(rootId.Clone().Dot(field.Name).Call(args))
								}
							})
					}
				}))
		}
	}
	return fmt.Sprintf("%#v", gen)
}

func typeCode(typ types.Type) jen.Code {
	return nil
}
