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
	"testing"

	"github.com/goplus/gop/token"
)

func newEmptySig() *types.Signature {
	return types.NewSignatureType(nil, nil, nil, nil, nil, false)
}

func TestBasic(t *testing.T) {
	pkg := NewPackage("", "foo", nil)
	pkg.StartFunc("main", newEmptySig(), 0)
	pkg.Stmts(
		pkg.StartIf().
			Init(pkg.DefineVar("v", 10, false)).
			Cond(pkg.BinaryOp(pkg.VarVal("v"), token.LSS, 50)),
		pkg.Stmts(
			pkg.Call(false, "println", "Hello, world!"),
		),
		pkg.End(),
	)
	pkg.End()
}
