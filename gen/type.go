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

type Kind uint

const (
	Invalid Kind = 1 << iota
	Basic
	String
	Pointer
	Slice
	Map
	Signature
	Struct
	Interface
	Chan
	Array
)

func typeKind(t types.Type) Kind {
retry:
	switch v := t.(type) {
	case *types.Basic:
		if (v.Info() & types.IsString) != 0 {
			return String
		}
		return Basic
	case *types.Pointer:
		return Pointer
	case *types.Slice:
		return Slice
	case *types.Map:
		return Map
	case *types.Signature:
		return Signature
	case *types.Struct:
		return Struct
	case *types.Interface:
		return Interface
	case *types.Named:
		t = v.Underlying()
		goto retry
	case *types.Chan:
		return Chan
	case *types.Array:
		return Array
	default:
		return Invalid
	}
}

func assertType(V, T types.Type) bool {
	if T == nil {
		return false
	}
	return types.AssignableTo(V, T)
}

// -----------------------------------------------------------------------------

type Type struct {
	types.Type
}

func (t Type) String() {} // disable func String

func (t Type) IsNil() bool {
	return t.Type == nil
}

// Kind returns the specific kind of this type.
func (t Type) Kind() Kind {
	return typeKind(t.Type)
}

// Key returns a map type's key type.
// It panics if the type's Kind is not Map.
func (t Type) Key() Type {
	panic("todo")
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

type callTuple struct {
	types.Tuple
	x *Expr
}

func NewExprs(list []*Expr) *Exprs {
	return &Exprs{list}
}

func TupleOf(list []*Expr) Tuple {
	if len(list) == 1 {
		if t, ok := list[0].typ.(*types.Tuple); ok {
			return &callTuple{*t, list[0]}
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

func (p *Package) checkAssignable(lht *Exprs, rht Tuple, org ast.Node) {
	nl, nr := lht.Len(), rht.Len()
	if nl != nr {
		if ct, ok := rht.(*callTuple); ok {
			p.errorf(org, "assignment mismatch: %d variables but %v returns %d values", nl, ct.x.Caller(), nr)
		} else {
			p.errorf(org, "assignment mismatch: %d variables but %d values", nl, nr)
		}
		return
	}
	for i := 0; i < nr; i++ {
		tl, tr := lht.At(i), rht.At(i)
		t, ok := tl.Type().(*RefType)
		if !ok {
			lhe := lht.list[i]
			p.error(lhe.org, "lhs expression $(code) is unassignable")
			return
		}
		ttr := tr.Type()
		if !types.AssignableTo(ttr, t.Type) {
			p.errorf(org, "assignment mismatch: can't assign type %v to %v", ttr, t.Type)
			return
		}
	}
}

// -----------------------------------------------------------------------------
