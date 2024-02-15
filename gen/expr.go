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

func (v *Expr) Stmt(pkg *Package) ast.Stmt {
	panic("todo")
}

// Kind returns v's Kind.
func (v *Expr) Kind() Kind {
	return typeKind(v.typ)
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

// lem returns the value that the expr v contains or that the pointer v
// points to. It panics if v's Kind is not a Pointer.
func (p *Package) Elem(v *Expr, org ...ast.Node) *Expr {
	t := v.typ.(*types.Pointer)
	return &Expr{
		src: &ast.StarExpr{X: v.src},
		typ: t.Elem(),
		org: origin(org),
	}
}

// Index func:
//   - a[i], a[key]
//   - fn[T1, T2, ..., Tn]
//   - G[T1, T2, ..., Tn]
//
// If v is a map, Index returns the value associated with key in the map v.
// If v is a Slice, String or Array, Index returns v's i'th element.
// If v is a generic function or class, instantiate it.
func (p *Package) Index(twoValue bool, v *Expr, args ...*Expr) *Expr {
	var typ = v.typ
	var key, elem types.Type
retry:
	switch t := typ.(type) {
	case *types.Slice:
		key, elem = types.Typ[types.Int], t.Elem()
	case *types.Map:
		key, elem = t.Key(), t.Elem()
		if twoValue {
			elem = types.NewTuple(
				types.NewVar(token.NoPos, p.Types, "", elem),
				types.NewVar(token.NoPos, p.Types, "", types.Typ[types.Bool]))
		}
	case *types.Basic:
		if info := t.Info(); (info & types.IsString) != 0 {
			key, elem = types.Typ[types.Int], types.Typ[types.Byte]
		}
	case *types.Array:
		key, elem = types.Typ[types.Int], t.Elem()
	case *types.Named:
		typ = t.Underlying()
		goto retry
	default:
		panic("todo - generic")
	}
	if elem == nil || len(args) != 1 {
		panic("todo")
	}
	i := args[0]
	if !assertType(i.typ, key) {
		elem = types.Typ[types.Invalid]
	}
	return &Expr{
		src: &ast.IndexExpr{X: v.src, Index: i.src},
		typ: elem,
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
