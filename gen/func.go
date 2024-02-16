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

type Func struct {
	obj   *types.Func
	name  string
	body  []Stmt
	doc   *ast.CommentGroup
	org   ast.Node
	flags FuncFlags
}

func (p *Func) Decl(pkg *Package) *ast.FuncDecl {
	obj := p.obj
	ctx := pkg.NewCtx(obj.Scope())
	flags := p.flags
	ret := &ast.FuncDecl{
		Doc:      p.doc,
		Name:     ast.NewIdent(p.name),
		Body:     makeBlockStmt(ctx, p.body),
		Operator: (flags & FuncIsOperator) != 0,
		Shadow:   (flags & FuncIsShadow) != 0,
		IsClass:  (flags & FuncIsClass) != 0,
	}
	sig := obj.Type().(*types.Signature)
	pkg.toFuncDecl(ret, sig)
	return ret
}

func (p *Func) Lit(pkg *Package) *ast.FuncLit {
	obj := p.obj
	ctx := pkg.NewCtx(obj.Scope())
	ret := &ast.FuncLit{
		Type: &ast.FuncType{},
		Body: makeBlockStmt(ctx, p.body),
	}
	sig := obj.Type().(*types.Signature)
	pkg.toFuncType(ret.Type, sig)
	return ret
}

func (p *Func) Doc(doc *ast.CommentGroup) *Func {
	p.doc = doc
	return p
}

func (p *Func) Body(list ...Stmt) *Func {
	p.body = list
	return p
}

func (p *Func) BodyAdd(list ...Stmt) *Func {
	p.body = append(p.body, list...)
	return p
}

// -----------------------------------------------------------------------------

type FuncFlags int

const (
	FuncIsOperator FuncFlags = 1 << iota
	FuncIsShadow
	FuncIsClass
)

func (p *Package) NewFunc(name string, sig *types.Signature, flags FuncFlags, org ...ast.Node) *Func {
	typs := p.Types
	obj := types.NewFunc(token.NoPos, typs, name, sig)
	if old := typs.Scope().Insert(obj); old != nil {
		panic("todo")
	}
	return &Func{obj, name, nil, nil, origin(org), flags}
}

func (p *Package) FuncLit(sig *types.Signature, org ...ast.Node) *Func {
	obj := types.NewFunc(token.NoPos, p.Types, "", sig)
	return &Func{obj, "", nil, nil, origin(org), 0}
}

// -----------------------------------------------------------------------------
