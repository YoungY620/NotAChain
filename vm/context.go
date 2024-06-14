package vm

// Context 保存执行环境中的所有状态
type Context struct {
	Memory *Memory  // 内存，用于存储变量和程序状态
	Stack  []uint64 // 栈，用于保存临时数据和操作数
	PC     int      // 程序计数器，用于控制程序的执行顺序
}

// NewContext 创建并初始化一个新的执行环境
func NewContext(cell []byte) *Context {
	return &Context{
		Memory: NewMemory(cell),
		Stack:  make([]uint64, 0),
		PC:     0,
	}
}

// Push 将值压入栈顶
func (c *Context) Push(value uint64) {
	c.Stack = append(c.Stack, value)
}

// Pop 从栈顶弹出值
func (c *Context) Pop() uint64 {
	if len(c.Stack) == 0 {
		panic("pop from an empty stack")
	}
	value := c.Stack[len(c.Stack)-1]
	c.Stack = c.Stack[:len(c.Stack)-1]
	return value
}

// Peek 从栈顶取值但不弹出
func (c *Context) Peek() uint64 {
	if len(c.Stack) == 0 {
		panic("peek from an empty stack")
	}
	return c.Stack[len(c.Stack)-1]
}

// IncrementPC 增加程序计数器
func (c *Context) IncrementPC() {
	c.PC++
}

// SetPC 设置程序计数器的值
func (c *Context) SetPC(newPC int) {
	c.PC = newPC
}
