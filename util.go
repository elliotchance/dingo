package main

import (
	"go/ast"
)

func newIdent(name string) *ast.Ident {
	return &ast.Ident{
		Name: name,
	}
}

func newFieldList(values ...string) *ast.FieldList {
	fields := []*ast.Field{}

	for _, value := range values {
		fields = append(fields, &ast.Field{
			Type: newIdent(value),
		})
	}

	return &ast.FieldList{
		List: fields,
	}
}

func newFunc(name string, params []string, returns []string, body *ast.BlockStmt) *ast.FuncDecl {
	return &ast.FuncDecl{
		Name: newIdent(name),
		Type: &ast.FuncType{
			Params:  newFieldList(params...),
			Results: newFieldList(returns...),
		},
		Body: body,
	}
}

func newReturn(expressions ...ast.Expr) *ast.ReturnStmt {
	var results []ast.Expr

	for _, expr := range expressions {
		results = append(results, expr)
	}

	return &ast.ReturnStmt{
		Results: results,
	}
}

func newBlock(stmts ...ast.Stmt) *ast.BlockStmt {
	return &ast.BlockStmt{
		List: stmts,
	}
}

func newCompositeLit(ty string, m map[string]ast.Expr) *ast.CompositeLit {
	var exprs []ast.Expr
	var keys []string

	for k := range m {
		keys = append(keys, k)
	}

	for _, k := range keys {
		exprs = append(exprs, &ast.KeyValueExpr{
			Key:   newIdent(k),
			Value: m[k],
		})
	}

	return &ast.CompositeLit{
		Type: newIdent(ty),
		Elts: exprs,
	}
}
