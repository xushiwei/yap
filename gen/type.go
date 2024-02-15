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
	"fmt"
	"go/types"

	"github.com/goplus/gop/ast"
)

// -----------------------------------------------------------------------------

type Type struct {
	types.Type
}

func (t Type) IsNil() bool {
	return t.Type == nil
}

// Elem returns a type's element type.
// It panics if the type's Kind is not String, Array, Chan, Map, Pointer, or Slice.
func (t Type) Elem() Type {
	panic("todo")
}

func (t Type) Slice() Type {
	return Type{types.NewSlice(t.Type)}
}

var (
	TyNil  = Type{}
	TyInt  = Type{types.Typ[types.Int]}
	TyByte = Type{types.Typ[types.Byte]}
	TyBool = Type{types.Typ[types.Bool]}
	TyAny  = Type{types.NewInterfaceType(nil, nil)}
)

// -----------------------------------------------------------------------------

type TypeEx interface {
	types.Type
	typeEx()
}

// -----------------------------------------------------------------------------

type RefType struct {
	Type types.Type
}

func NewRefType(typ types.Type) *RefType {
	return &RefType{typ}
}

func Assignable(lht, rht types.Type) bool {
	panic("todo")
}

func (p *RefType) Underlying() types.Type { return p }
func (p *RefType) String() string         { return fmt.Sprintf("RefType{%v}", p.Type) }
func (p *RefType) typeEx()                {}

// -----------------------------------------------------------------------------

type TypeType struct {
	Type types.Type
}

func NewTypeType(typ types.Type) *TypeType {
	return &TypeType{typ}
}

func (p *TypeType) Underlying() types.Type { return p }
func (p *TypeType) String() string         { return fmt.Sprintf("TypeType{%v}", p.Type) }
func (p *TypeType) typeEx()                {}

// -----------------------------------------------------------------------------

type Tuple interface {
	Len() int
	At(i int) *types.Var
}

type Exprs struct {
	list []*Expr
}

func NewExprs(list []*Expr) *Exprs {
	return &Exprs{list}
}

func TupleOf(list []*Expr) Tuple {
	if len(list) == 1 {
		if t, ok := list[0].typ.(*types.Tuple); ok {
			return t
		}
	}
	return NewExprs(list)
}

func (p *Exprs) Len() int {
	return len(p.list)
}

func (p *Exprs) At(i int) *types.Var {
	v := p.list[i]
	return types.NewVar(originPos(v.org), nil, "", v.typ)
}

// -----------------------------------------------------------------------------

func (p *Package) CheckAssignable(lht, rht Tuple, org ast.Node) (err error) {
	nl, nr := lht.Len(), rht.Len()
	if nl != nr {
		panic("todo")
	}
	for i := 0; i < nl; i++ {
		tl, tr := lht.At(i), rht.At(i)
		if !Assignable(tl.Type(), tr.Type()) {
			panic("todo")
		}
	}
	return nil
}

// -----------------------------------------------------------------------------
