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
	"go/token"
	"go/types"

	"github.com/goplus/gop/ast"
)

// -----------------------------------------------------------------------------

type BlockCtx struct {
	*Package
	Scope *types.Scope

	parent *BlockCtx
}

func (p *BlockCtx) New(comment string, org ast.Node) *BlockCtx {
	var start, end token.Pos
	if org != nil {
		start, end = org.Pos(), org.End()
	}
	scope := types.NewScope(p.Scope, start, end, comment)
	return &BlockCtx{p.Package, scope, p}
}

// -----------------------------------------------------------------------------

func (p *Package) NewCtx(scope *types.Scope) *BlockCtx {
	return &BlockCtx{p, scope, nil}
}

// -----------------------------------------------------------------------------
