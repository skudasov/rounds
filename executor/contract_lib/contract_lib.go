package contract_lib

import (
	"fmt"
	"github.com/yuin/gopher-lua"
)

func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.SetField(mod, "name", lua.LString("value"))
	L.Push(mod)
	return 1
}

var exports = map[string]lua.LGFunction{
	"call": call,
}

func call(L *lua.LState) int {
	fmt.Printf("contract function called")
	// call contract here and push data on stack
	L.Push(lua.LString("contract result"))
	return 1
}
