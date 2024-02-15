/*
 * Copyright (c) 2024 The GoPlus Authors (goplus.org). All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package gen

import (
	"go/types"

	"github.com/goplus/gop/ast"
	"github.com/goplus/gop/token"
)

// -----------------------------------------------------------------------------

type Var struct {
	name *ast.Ident
	typ  types.Type
	src  ast.Node
	org  ast.Node
	auto bool
}

func (p *Var) Stmt() ast.Stmt {
	switch v := p.src.(type) {
	case *ast.AssignStmt:
		return v
	case *ast.GenDecl:
		return &ast.DeclStmt{Decl: v}
	}
	panic("(*Var).Stmt: unexpected")
}

func (p *Var) Ref(org ...ast.Node) *Expr {
	return &Expr{
		src: p.name,
		typ: NewRefType(p.typ),
		org: origin(org),
	}
}

func (p *Var) Val(org ...ast.Node) *Expr {
	return &Expr{
		src: p.name,
		typ: p.typ,
		org: origin(org),
	}
}

// -----------------------------------------------------------------------------

func (p *Package) DefineVar(name string, val any, auto bool, org ...ast.Node) *Var {
	v := p.NewExpr(val)
	id := ast.NewIdent(name)
	return &Var{
		name: id,
		typ:  v.typ,
		src: &ast.AssignStmt{
			Lhs: []ast.Expr{id},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{v.src},
		},
		org:  origin(org),
		auto: auto,
	}
}

func (p *Package) NewVar(name string, typ Type, val any, auto bool, org ...ast.Node) *Var {
	if typ.Type == nil && val == nil {
		panic("NewVar: typ and val should not both be nil")
	}
	id := ast.NewIdent(name)
	spec := &ast.ValueSpec{Names: []*ast.Ident{id}}
	var expr *Expr
	if val != nil {
		expr = p.NewExpr(val)
		spec.Values = []ast.Expr{expr.src}
	}
	if typ.Type != nil {
		spec.Type = p.TypeExpr(typ.Type)
	} else {
		typ.Type = expr.typ
	}
	return &Var{
		name: id,
		typ:  typ.Type,
		src: &ast.GenDecl{
			Tok:   token.VAR,
			Specs: []ast.Spec{spec},
		},
		org:  origin(org),
		auto: auto,
	}
}

// v = append(v, arg)
func (p *Package) Append(v *Var, arg any, org ...ast.Node) *AssignStmt {
	rhs := p.Call("append", v.Val(), arg)
	return p.Assign(org...).Lhs(v.Ref()).Rhs(rhs)
}

// -----------------------------------------------------------------------------
