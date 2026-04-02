package engine

// Context 注册函数的上下文
// 用于在注册的 Go 函数中访问引擎状态和操作值
type Context struct {
	engine *Engine
	vm     *VM
}

// NewContext 创建新的上下文
func NewContext(engine *Engine, vm *VM) *Context {
	return &Context{
		engine: engine,
		vm:     vm,
	}
}

// Engine 返回关联的引擎实例
func (c *Context) Engine() *Engine {
	return c.engine
}

// VM 返回关联的虚拟机实例
func (c *Context) VM() *VM {
	return c.vm
}

// ThrowNewError 抛出新的运行时错误
func (c *Context) ThrowNewError(message string) error {
	return NewRuntimeError(message)
}

// ThrowNewArgError 抛出参数错误
func (c *Context) ThrowNewArgError(message string) error {
	return NewEngineError(message)
}

// NewResult 创建返回值
func (c *Context) NewResult(value Value) Value {
	return value
}

// ResultNull 返回 null 结果
func (c *Context) ResultNull() Value {
	return NewNull()
}

// ResultBool 返回布尔结果
func (c *Context) ResultBool(v bool) Value {
	return NewBool(v)
}

// ResultInt 返回整数结果
func (c *Context) ResultInt(v int64) Value {
	return NewInt(v)
}

// ResultFloat 返回浮点结果
func (c *Context) ResultFloat(v float64) Value {
	return NewFloat(v)
}

// ResultString 返回字符串结果
func (c *Context) ResultString(v string) Value {
	return NewString(v)
}

// ResultArray 返回数组结果
func (c *Context) ResultArray(v []Value) Value {
	return NewArray(v)
}

// ResultObject 返回对象结果
func (c *Context) ResultObject(v map[string]Value) Value {
	return NewObject(v)
}
