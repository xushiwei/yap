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
	"github.com/goplus/gop/ast"
	"github.com/goplus/gop/token"
)

type Stmt interface {
	Stmt() ast.Stmt
}

// -----------------------------------------------------------------------------

type AssignStmt struct {
	lhs []*Expr
	rhs []*Expr
	org ast.Node
	pkg *Package
}

func (p *Package) Assign(org ...ast.Node) *AssignStmt {
	return &AssignStmt{org: origin(org), pkg: p}
}

func (p *AssignStmt) Stmt() ast.Stmt {
	lht := NewExprs(p.lhs)
	rht := TupleOf(p.rhs)
	err := p.pkg.CheckAssignable(lht, rht, p.org)
	if err != nil {
		panic(err)
	}
	return &ast.AssignStmt{
		Lhs: nil,
		Tok: token.ASSIGN,
		Rhs: nil,
	}
}

func (p *AssignStmt) Lhs(lhs ...*Expr) *AssignStmt {
	p.lhs = lhs
	return p
}

func (p *AssignStmt) Rhs(rhs ...*Expr) *AssignStmt {
	p.rhs = rhs
	return p
}

// -----------------------------------------------------------------------------

type BlockStmt struct {
	body []Stmt
}

func (p *Package) Block(in ...Stmt) *BlockStmt {
	return &BlockStmt{in}
}

func (p *BlockStmt) Stmt() ast.Stmt {
	return makeBlockStmt(p.body)
}

func (p *BlockStmt) Add(list ...Stmt) *BlockStmt {
	p.body = append(p.body, list...)
	return p
}

func makeBlockStmt(in []Stmt) *ast.BlockStmt {
	list := make([]ast.Stmt, len(in))
	for i, v := range in {
		list[i] = v.Stmt()
	}
	return &ast.BlockStmt{List: list}
}

// -----------------------------------------------------------------------------

type IfStmt struct {
	init  Stmt
	cond  *Expr
	body  []Stmt
	else_ Stmt
	org   ast.Node
}

func (p *Package) If(org ...ast.Node) *IfStmt {
	return &IfStmt{org: origin(org)}
}

func (p *IfStmt) Stmt() ast.Stmt {
	stmt := &ast.IfStmt{
		Cond: p.cond.src,
		Body: makeBlockStmt(p.body),
	}
	if p.init != nil {
		stmt.Init = p.init.Stmt()
	}
	if p.else_ != nil {
		stmt.Else = p.else_.Stmt()
	}
	return stmt
}

func (p *IfStmt) Init(init Stmt) *IfStmt {
	p.init = init
	return p
}

func (p *IfStmt) Cond(cond *Expr) *IfStmt {
	p.cond = cond
	return p
}

func (p *IfStmt) Body(list ...Stmt) *IfStmt {
	p.body = list
	return p
}

func (p *IfStmt) Add(list ...Stmt) *IfStmt {
	p.body = append(p.body, list...)
	return p
}

func (p *IfStmt) Else(list ...Stmt) *IfStmt {
	switch len(list) {
	case 0:
		p.else_ = nil
	case 1:
		switch v := list[0].(type) {
		case *IfStmt:
			p.else_ = v
			return p
		case *BlockStmt:
			p.else_ = v
			return p
		}
		fallthrough
	default:
		p.else_ = &BlockStmt{list}
	}
	return p
}

// -----------------------------------------------------------------------------

type ForStmt struct {
	init Stmt // initialization statement; or nil
	cond any  // condition; or nil
	post Stmt // post iteration statement; or nil
	body []Stmt
	org  ast.Node
}

func (p *Package) For(org ...ast.Node) *ForStmt {
	return &ForStmt{org: origin(org)}
}

func (p *ForStmt) Stmt() ast.Stmt {
	stmt := &ast.ForStmt{
		Body: makeBlockStmt(p.body),
	}
	if p.init != nil {
		stmt.Init = p.init.Stmt()
	}
	if p.cond != nil {
		stmt.Cond = p.cond.(*Expr).src
	}
	if p.post != nil {
		stmt.Post = p.post.Stmt()
	}
	return stmt
}

func (p *ForStmt) Init(init Stmt) *ForStmt {
	p.init = init
	return p
}

func (p *ForStmt) Cond(cond *Expr) *ForStmt {
	cond.AssertType(TyBool)
	p.cond = cond
	return p
}

func (p *ForStmt) Post(post Stmt) *ForStmt {
	p.post = post
	return p
}

func (p *ForStmt) Body(list ...Stmt) *ForStmt {
	p.body = list
	return p
}

func (p *ForStmt) Add(list ...Stmt) *ForStmt {
	p.body = append(p.body, list...)
	return p
}

// -----------------------------------------------------------------------------

type RangeStmt struct {
	key, val *Var // Key, Value may be nil
	x        *Expr
	body     []Stmt
	org      ast.Node
}

func (p *Package) ForEach(k, v string, auto bool, x *Expr, org ...ast.Node) (loop *RangeStmt, key, val *Var) {
	switch x.Kind() {
	case Slice:
		key = rangeVar(k, TyInt, auto)
		if v != "" {
			val = rangeVar(v, x.Type().Elem(), auto)
		}
	default:
		panic("todo")
	}
	loop = &RangeStmt{key, val, x, nil, origin(org)}
	return
}

func rangeVar(name string, typ Type, auto bool) *Var {
	if name == "" {
		return nil
	}
	return &Var{ast.NewIdent(name), typ.Type, nil, nil, auto}
}

func rangeVarRef(v *Var) ast.Expr {
	if v == nil {
		return nil
	}
	return v.name
}

func (p *RangeStmt) Stmt() ast.Stmt {
	return &ast.RangeStmt{
		Key:   rangeVarRef(p.key),
		Value: rangeVarRef(p.val),
		Tok:   token.DEFINE,
		X:     p.x.src,
		Body:  makeBlockStmt(p.body),
	}
}

func (p *RangeStmt) Body(list ...Stmt) *RangeStmt {
	p.body = list
	return p
}

// -----------------------------------------------------------------------------
