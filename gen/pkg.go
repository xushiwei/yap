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

func origin(org []ast.Node) ast.Node {
	if org != nil {
		return org[0]
	}
	return nil
}

func originPos(org ast.Node) token.Pos {
	if org != nil {
		return org.Pos()
	}
	return token.NoPos
}

// -----------------------------------------------------------------------------

type PkgRef struct {
	Types *types.Type
}

func (p PkgRef) Ref(name string) Expr {
	panic("todo")
}

// -----------------------------------------------------------------------------

type Package struct {
}

func (p *Package) Import(pkgPath string) PkgRef {
	panic("todo")
}

func (p *Package) Typ(typ Type) *Expr {
	return &Expr{p.TypeExpr(typ.Type), NewTypeType(typ.Type), nil, nil}
}

func (p *Package) TypeExpr(typ types.Type) ast.Expr {
	panic("todo")
}

// -----------------------------------------------------------------------------
