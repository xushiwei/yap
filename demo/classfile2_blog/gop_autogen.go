// Code generated by gop (Go+); DO NOT EDIT.

package main

import "github.com/goplus/yap"

const _ = true

type get struct {
	yap.Handler
	*AppV2
}
type get_p_id struct {
	yap.Handler
	*AppV2
}
type AppV2 struct {
	yap.AppV2
}
//line demo/classfile2_blog/main.yap:1
func (this *AppV2) MainEntry() {
//line demo/classfile2_blog/main.yap:1:1
	this.Run(":8888")
}
func main() {
	yap.Gopt_AppV2_Main(new(AppV2), new(get), new(get_p_id))
}
//line demo/classfile2_blog/get.yap:1
func (this *get) Main(_gop_arg0 *yap.Context) {
//line demo/classfile2_blog/main.yap:1:1
	this.Handler.Main(_gop_arg0)
//line demo/classfile2_blog/get.yap:1:1
	this.Html__1(`<html><body>Hello, <a href="/p/123">YAP</a>!</body></html>`)
}
func (this *get) Classfname() string {
	return "get"
}
//line demo/classfile2_blog/get_p_#id.yap:1
func (this *get_p_id) Main(_gop_arg0 *yap.Context) {
	this.Handler.Main(_gop_arg0)
//line demo/classfile2_blog/get_p_#id.yap:1:1
	this.Yap__1("article", map[string]string{"id": this.Gop_Env("id")})
}
func (this *get_p_id) Classfname() string {
	return "get_p_#id"
}
