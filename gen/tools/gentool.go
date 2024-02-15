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

package tools

import (
	"errors"
	"flag"
	goast "go/ast"
	"go/types"
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"github.com/goplus/gop"
	"github.com/goplus/gop/ast"
	"github.com/goplus/gop/cl"
	"github.com/goplus/gop/parser"
	"github.com/goplus/gop/token"
	"github.com/goplus/gop/x/c2go"
	"github.com/goplus/gop/x/gopenv"
	"github.com/goplus/gop/x/typesutil"
	"github.com/goplus/mod/env"
	"github.com/goplus/mod/gopmod"
)

var (
	ErrSkip = errors.New("skip")
)

// -----------------------------------------------------------------------------

type Config struct {
	Gop      *env.Gop
	Fset     *token.FileSet
	Mod      *gopmod.Module
	Importer types.Importer
}

func NewDefaultConf(dir string) (conf *Config, err error) {
	mod, err := gop.LoadMod(dir)
	if err != nil {
		return
	}
	env := gopenv.Get()
	fset := token.NewFileSet()
	imp := gop.NewImporter(mod, env, fset)
	return &Config{Gop: env, Fset: fset, Mod: mod, Importer: imp}, nil
}

// -----------------------------------------------------------------------------

type Generator struct {
	self IGen

	onConf         func(dir string) (conf *Config, err error)
	onLoadAST      func(dir string) (ret []*ast.Package, err error)
	onASTOpened    func(pkg *ast.Package) error
	onTypesConf    func() *types.Config
	onNewTypes     func(pkg *ast.Package) *types.Package
	onNewTypesInfo func(pkg *ast.Package, typ *types.Package) (*types.Info, *typesutil.Info)
	onLoadTypes    func(pkg *Package, conf *types.Config) (err error)
	onTypesOpened  func(pkg *Package) error

	dirAbs string
	mod    *gopmod.Module
	fset   *token.FileSet
	imp    types.Importer
	filter func(fs.FileInfo) bool

	asts []*ast.Package
	typs []*Package

	finish bool
}

func (p *Generator) Finish() {
	p.finish = true
}

func (p *Generator) PkgPath() string {
	mod := p.mod
	rel := p.dirAbs[len(mod.Root()):]
	return mod.Path() + rel
}

func (p *Generator) Run() (err error) {
	conf, err := p.onConf(p.dirAbs)
	if err != nil || p.finish {
		return
	}
	p.fset = conf.Fset
	p.mod = conf.Mod
	p.imp = conf.Importer
	pkgs, err := p.onLoadAST(p.dirAbs)
	if err != nil || p.finish {
		return
	}
	if onOpened := p.onASTOpened; onOpened != nil {
		p.asts = make([]*ast.Package, 0, len(pkgs))
		for _, pkg := range pkgs {
			if err = onOpened(pkg); err != nil {
				if err != ErrSkip {
					return
				}
			} else {
				p.AddAST(pkg)
			}
			if p.finish {
				return
			}
		}
	} else {
		p.asts = pkgs
	}
	typesConf := p.onTypesConf()
	if p.finish {
		return
	}
	p.typs = make([]*Package, 0, len(p.asts))
	for _, ast := range p.asts {
		typ := p.onNewTypes(ast)
		if p.finish {
			return
		}
		goInfo, gopInfo := p.onNewTypesInfo(ast, typ)
		if p.finish {
			return
		}
		pkg := &Package{ast, typ, goInfo, gopInfo}
		err = p.onLoadTypes(pkg, typesConf)
		if err != nil || p.finish {
			return
		}
		if onOpened := p.onTypesOpened; onOpened != nil {
			err = onOpened(pkg)
			if err != nil && err != ErrSkip {
				return
			}
		}
		if p.finish {
			return
		}
		if err != ErrSkip {
			p.AddTypes(pkg)
		}
	}
	for _, typ := range p.typs {
		_ = typ
	}
	return nil
}

func (p *Generator) initGen(self IGen, dir string) {
	dirAbs, err := filepath.Abs(dir)
	if err != nil {
		log.Panicln("initGen:", err)
	}
	p.self = self
	p.dirAbs = dirAbs
	p.onConf = NewDefaultConf
	p.onLoadAST = p.LoadAST
	p.onTypesConf = p.DefaultTypesConf
	p.onNewTypes = p.NewDefaultTypes
	p.onNewTypesInfo = p.NewDefaultTypesInfo
	p.onLoadTypes = p.LoadTypes
}

func (p *Generator) OnConf(f func(dir string) (conf *Config, err error)) {
	p.onConf = f
}

func (p *Generator) OnFilter(f func(fi fs.FileInfo) bool) {
	p.filter = f
}

func (p *Generator) FilterNoTestFiles() {
	p.filter = gop.FilterNoTestFiles
}

func (p *Generator) OnLoadAST(f func(dir string) (ret []*ast.Package, err error)) {
	p.onLoadAST = f
}

func (p *Generator) OnASTOpened(f func(pkg *ast.Package) error) {
	p.onASTOpened = f
}

func (p *Generator) AddAST(pkg *ast.Package) {
	p.asts = append(p.asts, pkg)
}

func (p *Generator) OnTypesConf(f func() *types.Config) {
	p.onTypesConf = f
}

func (p *Generator) AddTypes(pkg *Package) {
	p.typs = append(p.typs, pkg)
}

// LoadAST is the default process of onLoadAST.
func (p *Generator) LoadAST(dir string) (ret []*ast.Package, err error) {
	ret = make([]*ast.Package, 2)
	ret[0], ret[1], err = p.ParseDir(dir)
	if err != nil {
		return
	}
	if ret[1] == nil {
		ret = ret[:1]
	}
	return
}

func (p *Generator) ParseDir(dir string) (out, test *ast.Package, err error) {
	pkgs, err := parser.ParseDirEx(p.fset, dir, parser.Config{
		ClassKind: p.mod.ClassKind,
		Filter:    p.filter,
		Mode:      parser.ParseComments | parser.SaveAbsFile,
	})
	if err != nil {
		return
	}
	for name, pkg := range pkgs {
		if strings.HasSuffix(name, "_test") {
			if test != nil {
				err = gop.ErrMultiTestPackges
				return
			}
			test = pkg
			continue
		}
		if out != nil {
			err = gop.ErrMultiPackges
			return
		}
		if len(pkg.Files) == 0 { // no Go+ source files
			continue
		}
		out = pkg
	}
	if out == nil {
		err = gop.ErrNotFound
	}
	return
}

func (p *Generator) DefaultTypesConf() *types.Config {
	return &types.Config{
		Context:  types.NewContext(),
		Importer: p.imp,
	}
}

func (p *Generator) NewDefaultTypes(pkg *ast.Package) *types.Package {
	return types.NewPackage("", pkg.Name)
}

func (p *Generator) NewDefaultTypesInfo(in *ast.Package, typ *types.Package) (*types.Info, *typesutil.Info) {
	return new(types.Info), new(typesutil.Info)
}

func (p *Generator) LoadTypes(pkg *Package, conf *types.Config) (err error) {
	fset := p.fset
	mod := p.mod
	ast := pkg.AST
	typ := pkg.Types
	_, err = cl.NewPackage(typ.Path(), ast, &cl.Config{
		Types:          typ,
		Fset:           fset,
		LookupPub:      c2go.LookupPub(mod),
		LookupClass:    mod.LookupClass,
		Importer:       conf.Importer,
		Recorder:       typesutil.NewRecorder(pkg.GopInfo),
		NoFileLine:     true,
		NoAutoGenMain:  true,
		NoSkipConstant: true,
	})
	if err != nil {
		return
	}
	if n := len(ast.GoFiles); n > 0 {
		files := make([]*goast.File, 0, n)
		for _, f := range ast.GoFiles {
			files = append(files, f)
		}
		scope := typ.Scope()
		objMap := typesutil.DeleteObjects(scope, files)
		checker := types.NewChecker(conf, fset, typ, pkg.GoInfo)
		err = checker.Files(files)
		typesutil.CorrectTypesInfo(scope, objMap, pkg.GopInfo.Uses)
	}
	return
}

// -----------------------------------------------------------------------------

type IGen interface {
	initGen(self IGen, dir string)
	Run() (err error)
}

func Gopt_Generator_Main(a IGen) {
	flag.Parse()
	dir := "."
	if flag.NArg() > 0 {
		dir = flag.Arg(0)
	}
	a.initGen(a, dir)
	if me, ok := a.(interface{ MainEntry() }); ok {
		me.MainEntry()
	}
	if err := a.Run(); err != nil {
		log.Panicln(err)
	}
}

// -----------------------------------------------------------------------------
