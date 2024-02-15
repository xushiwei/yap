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
	"go/constant"
	"go/types"

	"github.com/goplus/gop/ast"
	"github.com/goplus/gop/token"
)

// -----------------------------------------------------------------------------

type Expr struct {
	src ast.Expr
	typ types.Type
	val constant.Value
	org ast.Node
}

func (v *Expr) Origin(org ast.Node) *Expr {
	v.org = org
	return v
}

func (v *Expr) Stmt() ast.Stmt {
	panic("todo")
}

func (v *Expr) AssertKind(kind ...Kind) {
	panic("todo")
}

func (v *Expr) AssertType(t Type) {
	panic("todo")
}

// Type returns v's type.
func (v *Expr) Type() Type {
	return Type{v.typ}
}

// NumField returns the number of fields in the struct v.
// It panics if v's Kind is not Struct.
func (v *Expr) NumField() int {
	t := v.typ.Underlying().(*types.Struct)
	return t.NumFields()
}

// ToString returns string value of this expr.
// It panics if v's value isn't a constant string.
func (v *Expr) ToString() string {
	return constant.StringVal(v.val)
}

// -----------------------------------------------------------------------------

type Kind uint

const (
	Invalid Kind = iota
	Slice
	Array
	String
	Pointer
	Struct
)

func (v *Expr) Kind() Kind {
	return Invalid
}

// -----------------------------------------------------------------------------

// lem returns the value that the expr v contains or that the pointer v
// points to. It panics if v's Kind is not a Pointer.
func (v *Expr) Elem(org ...ast.Node) *Expr {
	t := v.typ.(*types.Pointer)
	return &Expr{
		src: &ast.StarExpr{X: v.src},
		typ: t.Elem(),
		org: origin(org),
	}
}

// Index returns v's i'th element. It panics if v's Kind is not Array, Slice, or
// String.
func (v *Expr) Index(i *Expr, org ...ast.Node) *Expr {
	i.AssertType(TyInt)
	v.AssertKind(Slice, String, Array)
	return &Expr{
		src: &ast.IndexExpr{X: v.src, Index: i.src},
		typ: v.Type().Elem(),
		org: origin(org),
	}
}

// -----------------------------------------------------------------------------

func (p *Package) NewExpr(val any) *Expr {
	switch v := val.(type) {
	case *Expr:
		return v
	}
	panic("todo")
}

func (p *Package) BinaryOp(x any, op token.Token, y any, org ...ast.Node) *Expr {
	panic("todo")
}

func (p *Package) UnaryOp(op token.Token, y *Expr, org ...ast.Node) *Expr {
	panic("todo")
}

func (p *Package) Call(fn any, args ...any) *Expr {
	panic("todo")
}

func (p *Package) CallSlice(fn any, args ...any) *Expr {
	panic("todo")
}

// -----------------------------------------------------------------------------

func (p *Package) Make(t Type, n any) *Expr {
	return p.Call("make", p.Typ(t), n)
}

func (p *Package) MakeCap(t Type, n, cap any) *Expr {
	return p.Call("make", p.Typ(t), n, cap)
}

// -----------------------------------------------------------------------------
