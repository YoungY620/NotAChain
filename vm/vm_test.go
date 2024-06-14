package vm

import (
	"neochain/common"
	"neochain/utils"
	"testing"
)

func TestEngine02(t *testing.T) {
	engine := NewVM(make([]byte, 1024))

	t.Logf("mem: %x", utils.LongBytesToInt(engine.Context.Memory.Cell))

	transaction := common.NewTransaction(0, 2)
	err := engine.ExecuteTransaction(transaction)
	if err != nil {
		t.Errorf(err.Error())
	}
	t.Logf("ret: %x", engine.Context.Peek())
	t.Logf("mem: %x", utils.LongBytesToInt(engine.Context.Memory.Cell))
}

func TestEngine21(t *testing.T) {
	engine := NewVM(make([]byte, 1024))

	t.Logf("mem: %x", utils.LongBytesToInt(engine.Context.Memory.Cell))

	transaction := common.NewTransaction(1, 3)
	err := engine.ExecuteTransaction(transaction)
	if err != nil {
		t.Errorf(err.Error())
	}
	t.Logf("ret: %x", engine.Context.Peek())
	t.Logf("mem: %x", utils.LongBytesToInt(engine.Context.Memory.Cell))
}
