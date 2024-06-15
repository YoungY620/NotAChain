package common

// Opcode 表示一个操作码和其参数
type Opcode struct {
	Name string
	Args []interface{}
}

type TxDefMsg struct {
	IdxFrom int `json:"idxFrom"`
	IdxTo   int `json:"idxTo"`
}

// Transaction 表示一个交易，包含一系列操作码
type Transaction struct {
	Code      []Opcode `json:"code"`
	RWSetHash string   `json:"rwSetHash"`
}

// NewTransaction 创建并初始化一个新的交易
func NewTransaction(idxFrom int, idxTo int) *Transaction {
	offsetFrom, offsetTo := idxFrom*8, idxTo*8
	return &Transaction{
		Code: []Opcode{
			{"LOAD", []interface{}{uint64(offsetFrom), uint64(8)}}, // 加载内存偏移量0处的8字节数据到堆栈
			{"DUP", nil},                         // 复制堆栈顶部元素
			{"PUSH", []interface{}{uint64(128)}}, // 将100推入堆栈
			{"CMP", nil},                         // 比较堆栈顶部两个元素
			{"SLEEP", nil},
			{"JEQ", []interface{}{9, uint64(0)}}, // 如果val<=100,跳转到第7条指令
			{"PUSH", []interface{}{uint64(2)}},   // 将2推入堆栈
			{"DIV", nil},                         // 将堆栈顶部元素除以2
			{"SLEEP", nil},
			{"JMP", []interface{}{uint64(11)}},  // 跳转到第9条指令
			{"PUSH", []interface{}{uint64(32)}}, // 将20推入堆栈
			{"ADD", nil},                        // 将堆栈顶部两个元素相加
			{"SLEEP", nil},
			{"STORE", []interface{}{uint64(offsetTo)}}, // 将结果存储回内存偏移量0处
		},
	}
}

type Block struct {
	Header BlockHeader `json:"header"`
	Txs    []TxDefMsg  `json:"txs"`
}

type BlockHeader struct {
	Height        int    `json:"height"`
	BlockHash     string `json:"blockHash"`
	PrevBlockHash string `json:"prevBlockHash"`
}

type RWSet struct {
	RSet map[string][]string `json:"rSet"`
	WSet map[string][]string `json:"wSet"`
}

type CommitMsg struct {
	Batch  []*TxDefMsg
	Height int
}
