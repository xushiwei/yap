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

package ydb

import (
	"context"
	"database/sql"
	"log"
	"reflect"
	"strings"

	"github.com/goplus/gop/token"
	"github.com/goplus/yap/gen"
)

// -----------------------------------------------------------------------------

// Insert inserts new rows.
//   - insert <colName1>, <val1>, <colName2>, <val2>, ...
//   - insert <colName1>, <valSlice1>, <colName2>, <valSlice2>, ...
//   - insert <structValOrPtr>
//   - insert <structOrPtrSlice>
func (p *Class) Insert(args ...any) (sql.Result, error) {
	if p.tbl == "" {
		log.Panicln("please call `use <tableName>` to specified current table")
	}
	nArg := len(args)
	if nArg == 1 {
		return p.insertStruc(args[0])
	}
	return p.insertKvPair(args...)
}

// Insert inserts new rows.
//   - insert <colName1>, <val1>, <colName2>, <val2>, ...
//   - insert <colName1>, <valSlice1>, <colName2>, <valSlice2>, ...
//   - insert <structValOrPtr>
//   - insert <structOrPtrSlice>
func (p *ClassGen) Insert(args ...Expr) Expr {
	if p.tbl == "" {
		log.Panicln("please call `use <tableName>` to specified current table")
	}
	nArg := len(args)
	if nArg == 1 {
		return p.insertStruc(args[0])
	}
	return p.insertKvPair(args...)
}

// Insert inserts a new row.
//   - insert <structValOrPtr>
//   - insert <structOrPtrSlice>
func (p *Class) insertStruc(arg any) (sql.Result, error) {
	vArg := reflect.ValueOf(arg)
	switch vArg.Kind() {
	case reflect.Slice:
		return p.insertStrucRows(vArg)
	case reflect.Pointer:
		vArg = vArg.Elem()
		fallthrough
	default:
		return p.insertStrucRow(vArg)
	}
}

// Insert inserts a new row.
//   - insert <structValOrPtr>
//   - insert <structOrPtrSlice>
func (p *ClassGen) insertStruc(arg Expr) Expr {
	switch arg.Kind() {
	case gen.Slice:
		return p.insertStrucRows(arg)
	case gen.Pointer:
		arg = arg.Elem()
		fallthrough
	default:
		return p.insertStrucRow(arg)
	}
}

func (p *Class) insertStrucRows(vSlice reflect.Value) (sql.Result, error) {
	rows := vSlice.Len()
	if rows == 0 {
		return nil, nil
	}
	hasPtr := false
	elem := vSlice.Type().Elem()
	kind := elem.Kind()
	if kind == reflect.Pointer {
		elem, hasPtr = elem.Elem(), true
		kind = elem.Kind()
	}
	if kind != reflect.Struct {
		log.Panicln("usage: insert <structOrPtrSlice>")
	}
	n := elem.NumField()
	names, cols := getCols(make([]string, 0, n), make([]field, 0, n), n, elem, 0)
	vals := make([]any, 0, len(names)*rows)
	for row := 0; row < rows; row++ {
		vElem := vSlice.Index(row)
		if hasPtr {
			vElem = vElem.Elem()
		}
		vals = getVals(vals, vElem, cols, true)
	}
	return p.insertRowsVals(p.tbl, names, vals, rows)
}

func (p *ClassGen) insertStrucRows(vSlice Expr) Expr {
	panic("todo")
}

func (p *Class) insertStrucRow(vArg reflect.Value) (sql.Result, error) {
	if vArg.Kind() != reflect.Struct {
		log.Panicln("usage: insert <structValOrPtr>")
	}
	n := vArg.NumField()
	names, cols := getCols(make([]string, 0, n), make([]field, 0, n), n, vArg.Type(), 0)
	vals := getVals(make([]any, 0, len(cols)), vArg, cols, true)
	return p.insertRow(p.tbl, names, vals)
}

func (p *ClassGen) insertStrucRow(vArg Expr) Expr {
	if vArg.Kind() != gen.Struct {
		log.Panicln("usage: insert <structValOrPtr>")
	}
	n := vArg.NumField()
	names, cols := getColsGen(make([]string, 0, n), make([]fieldGen, 0, n), n, vArg.Type(), 0)
	vals := getValsGen(make([]Expr, 0, len(cols)), vArg, cols, true)
	return p.insertRow(p.tbl, names, vals)
}

const (
	valFlagNormal  = 1
	valFlagSlice   = 2
	valFlagInvalid = valFlagNormal | valFlagSlice
)

// Insert inserts a new row.
//   - insert <colName1>, <val1>, <colName2>, <val2>, ...
//   - insert <colName1>, <valSlice1>, <colName2>, <valSlice2>, ...
func (p *Class) insertKvPair(kvPair ...any) (sql.Result, error) {
	nPair := len(kvPair)
	if nPair < 2 || nPair&1 != 0 {
		log.Panicln("usage: insert <colName1>, <val1>, <colName2>, <val2>, ...")
	}
	n := nPair >> 1
	names := make([]string, n)
	vals := make([]any, n)
	rows := -1      // slice length
	iArgSlice := -1 // -1: no slice, or index of first slice arg
	kind := 0
	for iPair := 0; iPair < nPair; iPair += 2 {
		i := iPair >> 1
		names[i] = kvPair[iPair].(string)
		val := kvPair[iPair+1]
		switch v := reflect.ValueOf(val); v.Kind() {
		case reflect.Slice:
			vlen := v.Len()
			if iArgSlice == -1 {
				iArgSlice = i
				rows = vlen
			} else if rows != vlen {
				log.Panicf("insert: unexpected slice length. got %d, expected %d\n", vlen, rows)
			} else {
				kind |= valFlagSlice
			}
			vals[i] = v
		default:
			kind |= valFlagNormal
			vals[i] = val
		}
	}
	if kind == valFlagInvalid {
		log.Panicln("insert: can't mix multiple slice arguments and normal value")
	}
	tbl := p.tblFromNames(names)
	if kind == valFlagSlice {
		return p.insertSlice(tbl, names, vals, rows)
	}
	if iArgSlice == -1 {
		return p.insertRow(tbl, names, vals)
	}
	return p.insertMulti(tbl, names, iArgSlice, vals)
}

// Insert inserts a new row.
//   - insert <colName1>, <val1>, <colName2>, <val2>, ...
//   - insert <colName1>, <valSlice1>, <colName2>, <valSlice2>, ...
func (p *ClassGen) insertKvPair(kvPair ...Expr) Expr {
	nPair := len(kvPair)
	if nPair < 2 || nPair&1 != 0 {
		log.Panicln("usage: insert <colName1>, <val1>, <colName2>, <val2>, ...")
	}
	n := nPair >> 1
	names := make([]string, n)
	vals := make([]Expr, n)
	iArgSlice := -1 // -1: no slice, or index of first slice arg
	kind := 0
	for iPair := 0; iPair < nPair; iPair += 2 {
		i := iPair >> 1
		names[i] = kvPair[iPair].ToString()
		v := kvPair[iPair+1]
		switch v.Kind() {
		case gen.Slice:
			if iArgSlice == -1 {
				iArgSlice = i
			} else {
				kind |= valFlagSlice
			}
		default:
			kind |= valFlagNormal
		}
		vals[i] = v
	}
	if kind == valFlagInvalid {
		log.Panicln("insert: can't mix multiple slice arguments and normal value")
	}
	tbl := p.tblFromNames(names)
	if kind == valFlagSlice {
		return p.insertSlice(tbl, names, vals)
	}
	if iArgSlice == -1 {
		return p.insertRow(tbl, names, vals)
	}
	return p.insertMulti(tbl, names, iArgSlice, vals)
}

// NOTE: len(args) == len(names)
func (p *Class) insertMulti(tbl string, names []string, iArgSlice int, args []any) (sql.Result, error) {
	argSlice := args[iArgSlice]
	defer func() {
		args[iArgSlice] = argSlice
	}()
	vArgSlice := argSlice.(reflect.Value)
	rows := vArgSlice.Len()
	vals := make([]any, 0, len(names)*rows)
	for i := 0; i < rows; i++ {
		args[iArgSlice] = vArgSlice.Index(i).Interface()
		vals = append(vals, args...)
	}
	return p.insertRowsVals(tbl, names, vals, rows)
}

func (p *ClassGen) insertMulti(tbl string, names []string, iArgSlice int, args []Expr) Expr {
	panic("todo")
}

// NOTE: len(args) == len(names)
func (p *Class) insertSlice(tbl string, names []string, args []any, rows int) (sql.Result, error) {
	vals := make([]any, 0, len(names)*rows)
	for i := 0; i < rows; i++ {
		for _, arg := range args {
			v := arg.(reflect.Value)
			vals = append(vals, v.Index(i).Interface())
		}
	}
	return p.insertRowsVals(tbl, names, vals, rows)
}

// NOTE: len(args) == len(names)
func (p *ClassGen) insertSlice(tbl string, names []string, args []Expr) Expr {
	pkg := p.Package
	vargs := make([]*gen.Var, len(names))
	for i, name := range names {
		vargs[i] = pkg.DefineVar(name, args[i], true)
	}
	log := pkg.Import("log")
	rows := pkg.DefineVar("rows", pkg.Call("len", vargs[0].Val()), true)
	body := pkg.Block(rows)
	for i := 1; i < len(vargs); i++ {
		vlen := pkg.DefineVar("vlen", pkg.Call("len", vargs[i].Val()), true)
		ifStmt := pkg.If().Init(vlen).Cond(pkg.BinaryOp(rows.Val(), token.NEQ, vlen.Val())).Body(
			pkg.Call(log.Ref("Panicf"), "insert: unexpected slice length. got %d, expected %d\n", vlen.Val(), rows.Val()),
		)
		body.Add(ifStmt)
	}

	// vals := make([]any, 0, len(names)*rows)
	//
	anySlice := pkg.Typ(gen.TyAny.Slice())
	n := pkg.BinaryOp(len(names), token.MUL, rows.Val())
	vals := pkg.DefineVar("vals", pkg.Call("make", anySlice, 0, n), true)

	// for i := 0; i < rows; i++ {
	//
	i := pkg.DefineVar("i", 0, true)
	cond := pkg.BinaryOp(i.Val(), token.LSS, rows.Val())
	inc := pkg.UnaryOp(token.INC, i.Ref())
	forStmt := pkg.For().Init(i).Cond(cond).Post(inc)
	for _, arg := range vargs {
		// vals = append(vals, v.Index(i))
		v := arg.Val()
		assign := pkg.Append(vals, v.Index(i.Val()))
		forStmt.Add(assign)
	}

	body.Add(vals, forStmt)
	return p.insertRowsVals(tbl, names, vals, rows)
}

// NOTE: len(vals) == len(names) * rows
func (p *Class) insertRowsVals(tbl string, names []string, vals []any, rows int) (sql.Result, error) {
	n := len(names)
	query := makeInsertExpr(tbl, names)
	query = append(query, valParams(n, rows)...)

	q := string(query)
	if debugExec {
		log.Println("==>", q, vals)
	}
	result, err := p.db.ExecContext(context.TODO(), q, vals...)
	return p.insertRet(result, err)
}

// NOTE: len(vals) == len(names) * rows
func (p *ClassGen) insertRowsVals(tbl string, names []string, vals, rows *gen.Var) Expr {
	n := len(names)
	query := makeInsertExpr(tbl, names)
	query = append(query, valParams(n, rows)...)

	q := string(query)
	if debugExec {
		log.Println("==>", q, vals)
	}
	result, err := p.db.ExecContext(context.TODO(), q, vals...)
	return p.insertRet(result, err)
}

func (p *Class) insertRow(tbl string, names []string, vals []any) (sql.Result, error) {
	if len(names) == 0 {
		log.Panicln("insert: nothing to insert")
	}
	query := makeInsertExpr(tbl, names)
	query = append(query, valParam(len(vals))...)

	q := string(query)
	if debugExec {
		log.Println("==>", q, vals)
	}
	result, err := p.db.ExecContext(context.TODO(), q, vals...)
	return p.insertRet(result, err)
}

func (p *ClassGen) insertRow(tbl string, names []string, vals []Expr) Expr {
	panic("todo")
}

func (p *Class) insertRet(result sql.Result, err error) (sql.Result, error) {
	if err != nil {
		p.handleErr("insert:", err)
	}
	return result, err
}

func makeInsertExpr(tbl string, names []string) []byte {
	query := make([]byte, 0, 128)
	query = append(query, "INSERT INTO "...)
	query = append(query, tbl...)
	query = append(query, ' ', '(')
	query = append(query, strings.Join(names, ",")...)
	query = append(query, ") VALUES "...)
	return query
}

func valParams(n, rows int) string {
	valparam := valParam(n)
	valparams := strings.Repeat(valparam+",", rows)
	valparams = valparams[:len(valparams)-1]
	return valparams
}

func valParamsGen(pkg *gen.Package, query string, n int, rows *gen.Var) string {
	valparam := valParam(n)

	// size := len(query) + (len(valparam)+1)*rows - 1
	mul := pkg.BinaryOp(len(valparam)+1, token.MUL, rows.Val())
	add := pkg.BinaryOp(len(query), token.ADD, mul)
	size := pkg.DefineVar("size", pkg.BinaryOp(add, token.SUB, 1), true)

	// b := make([]byte, 0, size)
	b := pkg.DefineVar("b", pkg.Call("make", gen.TyByte.Slice()))

	valparams := strings.Repeat(valparam+",", rows)
	valparams = valparams[:len(valparams)-1]
	return valparams
}

func valParam(n int) string {
	valparam := strings.Repeat("?,", n)
	valparam = "(" + valparam[:len(valparam)-1] + ")"
	return valparam
}

// -----------------------------------------------------------------------------
