package executor

import (
	mapper "github.com/yuin/gluamapper"
	lua "github.com/yuin/gopher-lua"
	"log"
)

type Executor struct {
	L *lua.LState
}

func NewExecutor() *Executor {
	return &Executor{L: lua.NewState()}
}

func (m *Executor) Load(filename string) (lua.LValue, error) {
	if err := m.L.DoFile(filename); err != nil {
		return nil, err
	}
	return m.L.Get(-1), nil
}

func (m *Executor) Execute(code string) (lua.LValue, error) {
	if err := m.L.DoString(code); err != nil {
		return nil, err
	}
	return m.L.Get(-1), nil
}

func (m *Executor) Call(fname string, nret int) ([]lua.LValue, error) {
	if err := m.L.CallByParam(lua.P{
		Fn:      m.L.GetGlobal(fname), // name of Lua function
		NRet:    nret,                 // number of returned values
		Protect: true,                 // return err or panic
	}); err != nil {
		return nil, err
	}

	results := make([]lua.LValue, 0)
	for i := 1; i <= nret; i++ {
		results = append(results, m.L.Get(-i))
	}

	//// Pop the returned value from the stack ?
	//m.L.Pop(1)
	return results, nil
}

type State struct {
	Name string
	Age  int
}

func (m *Executor) State() map[interface{}]interface{} {
	state := m.L.GetGlobal("State")
	if _, ok := state.(*lua.LTable); ok {
		var st map[interface{}]interface{}
		if err := mapper.Map(state.(*lua.LTable), &st); err != nil {
			log.Fatal(err)
		}
		return st
	}
	return nil
}

func (m *Executor) Shutdown() {
	m.L.Close()
}
