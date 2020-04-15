package executor

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	lua "github.com/yuin/gopher-lua"
	"log"
	"rounds/executor/contract_lib"
	"testing"
)

func TestNewExecutorConstructorStateDeserialized(t *testing.T) {
	e := NewExecutor()
	_, err := e.Load("test_contracts/global_class.lua")
	if err != nil {
		log.Fatal(err)
	}
	_, err = e.Execute(`State:new("John", 12, "subName")`)
	if err != nil {
		log.Fatal(err)
	}
	state := e.State()
	assert.Equal(t, state["Name"].(string), "John")
	assert.Equal(t, state["Age"].(float64), float64(12))
	assert.Equal(t, state["Child"].(map[interface{}]interface{})["Name"].(string), "subName")
}

func TestExecutorCallback(t *testing.T) {
	e := NewExecutor()
	ch := make(chan lua.LValue)
	e.L.SetGlobal("ch", lua.LChannel(ch))
	e.L.PreloadModule("contract", contract_lib.Loader)
	//_, err := e.Load("test_contracts/testlib.lua")
	//if err != nil {
	//	log.Fatal(err)
	//}
	_, err := e.Load("test_contracts/global_class_callback.lua")
	if err != nil {
		log.Fatal(err)
	}
	_, err = e.Execute(`State:new("John", 12, "subName")`)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		v := <-ch
		fmt.Printf("v from ch: %s\n", v)
	}()
	val, err := e.Call("getName", 2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("val: %s\n", val)
}
