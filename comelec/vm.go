package comelec

import (
	"github.com/robertkrimen/otto"
)

func NewComelecVm() (ctx *JsVmComelec) {
	vm := otto.New()

	ctx = new(JsVmComelec)
	ctx.vm = vm

	vm.Set("$include", ctx.jsFn_Include)
	vm.Set("$list", ctx.jsFn_List)
	vm.Set("$", ctx.jsFn_FromPath)
	vm.Set("$eachls", ctx.jsFn_EachList)
	vm.Set("$pretty", ctx.jsFn_PrettyPrn)
	vm.Set("$prn", ctx.jsFn_Print)

	vm.Set("$search", ctx.jsFn_Search)
	vm.Set("$kprefix", ctx.jsFn_FindPrefix)
	vm.Set("$update_index", ctx.jsFn_UpdateIndex)
	vm.Set("$dbg_raw", ctx.jsFn_DbgRaw)

	vm.Set("$libpath", ctx.jsFn_AddLibPath)
	vm.Set("$import", ctx.jsFn_ImportLib)

	return

}
