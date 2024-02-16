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
)

// -----------------------------------------------------------------------------

func (p *Package) toFuncType(ret *ast.FuncType, sig *types.Signature) {
}

func (p *Package) toFuncDecl(ret *ast.FuncDecl, sig *types.Signature) {
}

func (p *Package) TypeExpr(typ types.Type) ast.Expr {
	panic("todo")
}

// -----------------------------------------------------------------------------
