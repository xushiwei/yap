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

type Stmt interface {
	Stmt(ctx *BlockCtx) ast.Stmt
}

// -----------------------------------------------------------------------------

type AssignStmt struct {
	lhs []*Expr
	rhs []*Expr
	org ast.Node
}

func (p *Package) Assign(org ...ast.Node) *AssignStmt {
	return &AssignStmt{org: origin(org)}
}

func (p *AssignStmt) Stmt(ctx *BlockCtx) ast.Stmt {
	lht := NewExprs(p.lhs)
	rht := TupleOf(p.rhs)
	ctx.checkAssignable(lht, rht, p.org)
	return &ast.AssignStmt{
		Lhs: astExprs(p.lhs),
		Tok: token.ASSIGN,
		Rhs: astExprs(p.rhs),
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

func (p *BlockStmt) Stmt(ctx *BlockCtx) ast.Stmt {
	me := ctx.New("block stmt", nil)
	return makeBlockStmt(me, p.body)
}

func (p *BlockStmt) BodyAdd(list ...Stmt) *BlockStmt {
	p.body = append(p.body, list...)
	return p
}

func makeBlockStmt(ctx *BlockCtx, in []Stmt) *ast.BlockStmt {
	list := make([]ast.Stmt, len(in))
	for i, v := range in {
		list[i] = v.Stmt(ctx)
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

func (p *IfStmt) Stmt(ctx *BlockCtx) ast.Stmt {
	if !assertType(p.cond.typ, types.Typ[types.Bool]) {
		panic("todo")
	}
	ifCtx := ctx.New("if stmt", p.org)
	ifBody := ifCtx.New("if body")
	stmt := &ast.IfStmt{
		Cond: p.cond.src,
		Body: makeBlockStmt(ctx, p.body),
	}
	if p.init != nil {
		stmt.Init = p.init.Stmt(ifCtx)
	}
	if p.else_ != nil {
		stmt.Else = p.else_.Stmt(ctx)
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

func (p *IfStmt) BodyAdd(list ...Stmt) *IfStmt {
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
	init Stmt  // initialization statement; or nil
	cond *Expr // condition; or nil
	post Stmt  // post iteration statement; or nil
	body []Stmt
	org  ast.Node
}

func (p *Package) For(org ...ast.Node) *ForStmt {
	return &ForStmt{org: origin(org)}
}

func (p *ForStmt) Stmt(ctx *BlockCtx) ast.Stmt {
	stmt := &ast.ForStmt{
		Body: makeBlockStmt(ctx, p.body),
	}
	if p.init != nil {
		stmt.Init = p.init.Stmt(ctx)
	}
	if p.cond != nil {
		if !assertType(p.cond.typ, types.Typ[types.Bool]) {
			panic("todo")
		}
		stmt.Cond = p.cond.src
	}
	if p.post != nil {
		stmt.Post = p.post.Stmt(ctx)
	}
	return stmt
}

func (p *ForStmt) Init(init Stmt) *ForStmt {
	p.init = init
	return p
}

func (p *ForStmt) Cond(cond *Expr) *ForStmt {
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

func (p *ForStmt) BodyAdd(list ...Stmt) *ForStmt {
	p.body = append(p.body, list...)
	return p
}

// Times means
//   - for i := 0; i < n; i++
func Times(pkg *Package, n *Expr) (*ForStmt, *Var) {
	i := pkg.DefineVar("i", 0, true)
	cond := pkg.BinaryOp(i.Val(), token.LSS, n)
	inc := pkg.UnaryOp(token.INC, i.Ref())
	return pkg.For().Init(i).Cond(cond).Post(inc), i
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

func (p *RangeStmt) Stmt(ctx *BlockCtx) ast.Stmt {
	return &ast.RangeStmt{
		Key:   rangeVarRef(p.key),
		Value: rangeVarRef(p.val),
		Tok:   token.DEFINE,
		X:     p.x.src,
		Body:  makeBlockStmt(ctx, p.body),
	}
}

func (p *RangeStmt) Body(list ...Stmt) *RangeStmt {
	p.body = list
	return p
}

// -----------------------------------------------------------------------------
