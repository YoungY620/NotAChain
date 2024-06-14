package vm

import (
	"errors"
	"math/rand"
	"neochain/common"
	"neochain/utils"
	"time"
)

type OpAction = func(*Context, []interface{}) error

// VM 模拟一个简单的交易执行引擎
type VM struct {
	opcodeMap map[string]OpAction
	Context   *Context
}

// NewVM 创建并初始化一个新的执行引擎
func NewVM(cell []byte) *VM {
	e := &VM{
		Context:   NewContext(cell),
		opcodeMap: make(map[string]OpAction),
	}
	e.loadOpcodes()
	return e
}

// loadOpcodes 加载所有操作码及其对应的处理函数
func (e *VM) loadOpcodes() {
	e.opcodeMap["LOAD"] = load
	e.opcodeMap["STOREI"] = storei
	e.opcodeMap["STORE"] = store
	e.opcodeMap["MALLOC"] = malloc
	e.opcodeMap["ADD"] = add
	e.opcodeMap["SUB"] = sub
	e.opcodeMap["MUL"] = mul
	e.opcodeMap["DIV"] = div
	e.opcodeMap["CMP"] = cmp
	e.opcodeMap["JMP"] = jmp
	e.opcodeMap["JEQ"] = jeq
	e.opcodeMap["PUSH"] = push
	e.opcodeMap["DUP"] = dup
	e.opcodeMap["SLEEP"] = sleep
}

// ExecuteTransaction 执行给定的交易
func (e *VM) ExecuteTransaction(t *common.Transaction) error {
	e.Context.SetPC(0)
	for e.Context.PC < len(t.Code) {
		if e.Context.PC >= len(t.Code) {
			return errors.New("program counter out of bounds")
		}
		opcode := t.Code[e.Context.PC]
		//fmt.Printf("pc   : %x\n", e.Context.PC)
		//fmt.Printf("cmd  : %s(%s)\n", opcode.Name, opcode.Args)
		//fmt.Printf("mem  : %x\n", utils.LongBytesToInt(e.Context.Memory.Cell[:32]))
		//fmt.Printf("stack: %x\n", e.Context.Stack)
		if op, exists := e.opcodeMap[opcode.Name]; exists {
			err := op(e.Context, opcode.Args)
			if err != nil {
				return err
			}
		} else {
			return errors.New("unknown opcode")
		}
		e.Context.IncrementPC()
	}
	return nil
}

// 示例操作码函数
func load(ctx *Context, args []interface{}) error {
	offset := args[0].(uint64)
	length := args[1].(uint64)
	data, err := ctx.Memory.Map(offset, length)
	if err != nil {
		return err
	}
	datas := make([]uint64, 0)
	for i := 0; i < int(length); i += 8 {
		var endI int
		if uint64(i+8) > length {
			endI = int(length)
		} else {
			endI = i + 8
		}
		datas = append(datas, utils.BytesToInt(data[i:endI]))
	}
	for _, d := range datas {
		ctx.Push(d)
	}
	return nil
}

func storei(ctx *Context, args []interface{}) error {
	offset := args[0].(uint64)
	return ctx.Memory.Store(offset, args[1].([]byte))
}

func store(ctx *Context, args []interface{}) error {
	offset := args[0].(uint64)
	value := utils.UintToBytes(ctx.Peek())
	return ctx.Memory.Store(offset, value)
}

func malloc(ctx *Context, args []interface{}) error {
	size := args[0].(uint64)
	offset := uint64(ctx.Memory.Size())
	ctx.Memory.Malloc(offset, size)
	ctx.Push(offset)
	return nil
}

func add(ctx *Context, args []interface{}) error {
	a := ctx.Pop()
	b := ctx.Pop()
	ctx.Push(a + b)
	return nil
}

func sub(ctx *Context, args []interface{}) error {
	a := ctx.Pop()
	b := ctx.Pop()
	ctx.Push(b - a)
	return nil
}

func mul(ctx *Context, args []interface{}) error {
	a := ctx.Pop()
	b := ctx.Pop()
	ctx.Push(a * b)
	return nil
}

func div(ctx *Context, args []interface{}) error {
	a := ctx.Pop()
	b := ctx.Pop()
	if a == 0 {
		return errors.New("division by zero")
	}
	ctx.Push(b / a)
	return nil
}

func cmp(ctx *Context, args []interface{}) error {
	a := ctx.Pop()
	b := ctx.Pop()
	if a < b {
		ctx.Push(1)
	} else {
		ctx.Push(0)
	}
	return nil
}

func jmp(ctx *Context, args []interface{}) error {
	target := ctx.Peek()
	ctx.SetPC(int(target))
	return nil
}

func jeq(ctx *Context, args []interface{}) error {
	target := args[0].(int)
	a := args[1].(uint64)
	b := ctx.Peek()

	if a == b {
		ctx.SetPC(target)
	}

	return nil
}

func push(ctx *Context, args []interface{}) error {
	value := args[0].(uint64)
	ctx.Push(value)
	return nil
}

func dup(ctx *Context, args []interface{}) error {
	top := ctx.Peek()
	ctx.Push(top)
	return nil
}

func sleep(ctx *Context, args []interface{}) error {
	time.Sleep(time.Millisecond * time.Duration(300*rand.Float32()))
	return nil
}
