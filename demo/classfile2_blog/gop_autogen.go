// Code generated by gop (Go+); DO NOT EDIT.

package main

import "github.com/goplus/yap"

const _ = true

type get_p_id struct {
	yap.Handler
}

func main() {
	yap.Gopt_AppV2_Main(new(yap.AppV2), new(get_p_id))
}
//line demo/classfile2_blog/get_p_#id.yap:1
func (this *get_p_id) Main(_gop_arg0 *yap.Context) {
	this.Handler.Main(_gop_arg0)
//line demo/classfile2_blog/get_p_#id.yap:1:1
	this.Yap__1("article", map[string]string{"id": this.Param("id")})
}
func (this *get_p_id) Classfname() string {
	return "get_p_#id"
}
