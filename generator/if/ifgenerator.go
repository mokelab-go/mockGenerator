package ifgenerator

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"strings"

	gen "github.com/mokelab-go/mockGenerator/generator"
)

type generator struct {
}

func New() gen.Generator {
	return &generator{}
}

func (g *generator) Generate(src string) (string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return "", err
	}

	// ouput decls
	outDecls := make([]ast.Decl, 0)

	// comment
	cMap := ast.NewCommentMap(fset, f, f.Comments)

	// import
	imports := filterImportDecls(f)
	for _, n := range imports {
		outDecls = append(outDecls, n)
	}

	// find interface type
	decls := filterTypeDecls(f)
	for _, decl := range decls {
		if !isMockInterface(decl, cMap) {
			continue
		}
		for _, spec := range decl.Specs {
			if !isInterfaceTypeSpec(spec) {
				continue
			}

			typeSpec := spec.(*ast.TypeSpec)
			typeName := "Mock" + typeSpec.Name.Name
			iType := typeSpec.Type.(*ast.InterfaceType)
			fieldList := make([]*ast.Field, 0, len(iType.Methods.List))
			methodList := make([]*ast.FuncDecl, 0, len(iType.Methods.List))

			for _, method := range iType.Methods.List {
				methodName := method.Names[0].Name

				fType := method.Type.(*ast.FuncType)
				results := make([]ast.Expr, 0, len(fType.Results.List))
				for i := range fType.Results.List {
					results = append(results, ast.NewIdent(fmt.Sprintf("m.%sResult%d", methodName, i)))
				}

				methodList = append(methodList, &ast.FuncDecl{
					Name: ast.NewIdent(methodName),
					Recv: &ast.FieldList{
						List: []*ast.Field{
							&ast.Field{
								Names: []*ast.Ident{
									ast.NewIdent("m"),
								},
								Type: &ast.StarExpr{
									X: ast.NewIdent(typeName),
								},
							},
						},
					},
					Type: &ast.FuncType{
						Params:  fType.Params,
						Results: fType.Results,
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ReturnStmt{
								Results: results,
							},
						},
					},
				})

				for i, t := range fType.Results.List {
					fieldList = append(fieldList, &ast.Field{
						Names: []*ast.Ident{
							ast.NewIdent(fmt.Sprintf("%sResult%d", methodName, i)),
						},
						Type: t.Type,
					})
				}
			}
			structDecl := &ast.GenDecl{
				Tok: token.TYPE,
				Specs: []ast.Spec{
					&ast.TypeSpec{
						Name: ast.NewIdent(typeName),
						Type: &ast.StructType{
							Fields: &ast.FieldList{
								List: fieldList,
							},
						},
					},
				},
			}
			outDecls = append(outDecls, structDecl)
			for _, m := range methodList {
				outDecls = append(outDecls, m)
			}

		}
	}
	file := &ast.File{
		Name:  f.Name,
		Decls: outDecls,
	}

	buf := bytes.NewBuffer(make([]byte, 0, len(src)*3))
	format.Node(buf, fset, file)
	return buf.String(), nil
}

func filterImportDecls(f *ast.File) []*ast.GenDecl {
	out := make([]*ast.GenDecl, 0, len(f.Decls))
	for _, decl := range f.Decls {
		if isImportDec(decl) {
			out = append(out, decl.(*ast.GenDecl))
		}
	}
	return out
}

func filterTypeDecls(f *ast.File) []*ast.GenDecl {
	out := make([]*ast.GenDecl, 0, len(f.Decls))
	for _, decl := range f.Decls {
		if isTypeDec(decl) {
			out = append(out, decl.(*ast.GenDecl))
		}
	}
	return out
}

func isImportDec(decl ast.Decl) bool {
	decl2, ok := decl.(*ast.GenDecl)
	if !ok {
		return false
	}
	if decl2.Tok != token.IMPORT {
		return false
	}
	return true
}

func isTypeDec(decl ast.Decl) bool {
	decl2, ok := decl.(*ast.GenDecl)
	if !ok {
		return false
	}
	if decl2.Tok != token.TYPE {
		return false
	}
	return true
}

func isInterfaceTypeSpec(spec ast.Spec) bool {
	typeSpec, ok := spec.(*ast.TypeSpec)
	if !ok {
		return false
	}
	_, ok = typeSpec.Type.(*ast.InterfaceType)
	if !ok {
		return false
	}
	return true
}

func isMockInterface(decl *ast.GenDecl, cMap ast.CommentMap) bool {
	comments, ok := cMap[decl]
	if !ok {
		return false
	}
	for _, c := range comments {
		c2 := strings.SplitN(c.Text(), "\n", -1)
		for _, t := range c2 {
			if t == "+mock" {
				return true
			}
		}
	}
	return false
}
