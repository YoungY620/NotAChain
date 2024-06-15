package commit

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"log"
	"neochain/common"
	"neochain/utils"
	"neochain/vm"
	"sync"
	"time"
)

type Committer struct {
	BlockDB *leveldb.DB
	StateDB *leveldb.DB

	commitMutex sync.Mutex
	wmapMutex   sync.Mutex
}

func NewCommitter(chainID string) *Committer {
	blockDB, err := leveldb.OpenFile(fmt.Sprintf("tmp/my-raft-cluster/%s/blockdb.leveldb", chainID), nil)
	if err != nil {
		log.Fatalf("failed to open block db: %s", err)
	}
	stateDB, err := leveldb.OpenFile(fmt.Sprintf("tmp/my-raft-cluster/%s/statedb.leveldb", chainID), nil)
	if err != nil {
		log.Fatalf("failed to open state db: %s", err)
	}
	genesisBlock := &common.Block{
		Header: common.BlockHeader{
			Height:        0,
			BlockHash:     "",
			PrevBlockHash: "",
		},
		Txs: nil,
	}

	commiter := &Committer{
		BlockDB: blockDB,
		StateDB: stateDB,
	}
	commiter.storeBlock(0, genesisBlock, make([]byte, 1024))
	return commiter
}

func (c *Committer) CommitBlock(msg common.CommitMsg) []*common.TxDefMsg {
	log.Printf("performance statistic: exe[s][%d]: %v", msg.Height, time.Now())
	abortedTxs, successTxs, err, lastBlock, engine := c.executeBlock(msg)
	if err != nil {
		log.Fatalf("failed to execute block: %s", err)
	}
	log.Printf("performance statistic: exe[e][%d]: %v", msg.Height, time.Now())

	c.commitMutex.Lock()
	defer c.commitMutex.Unlock()

	log.Printf("performance statistic: commit[s][%d]: %v", msg.Height, time.Now())
	block := &common.Block{
		Header: common.BlockHeader{
			Height:        msg.Height,
			PrevBlockHash: lastBlock.Header.BlockHash,
			BlockHash:     "",
		},
		Txs: successTxs,
	}
	c.storeBlock(msg.Height, block, engine.Context.Memory.Cell)
	log.Printf("performance statistic: commit[e][%d]: %v", msg.Height, time.Now())

	return abortedTxs
}

func (c *Committer) storeBlock(height int, block *common.Block, state []byte) {
	curHeightBin := make([]byte, 8)
	binary.BigEndian.PutUint64(curHeightBin, uint64(height))
	blockBytes := calBlockHash(block)
	err := c.StateDB.Put(curHeightBin, state, nil)
	if err != nil {
		log.Fatalf("failed to put block: %s", err)
	}
	err = c.BlockDB.Put(curHeightBin, blockBytes, nil)
	if err != nil {
		log.Fatalf("failed to put block: %s", err)
	}
}

func calBlockHash(block *common.Block) []byte {
	blockBytes, err := json.Marshal(block)
	if err != nil {
		log.Fatalf("failed to marshal block: %s", err)
	}
	block.Header.BlockHash = hex.EncodeToString(calHash(blockBytes))
	return blockBytes
}

func calHash(blockBytes []byte) []byte {
	hash := sha256.New()
	hash.Write(blockBytes)
	return (hash.Sum(nil))
}

func calTxHash(uid int, txDef *common.TxDefMsg) []byte {
	uidBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(uidBytes, uint64(uid))
	txDefBytes, err := json.Marshal(*txDef)
	if err != nil {
		log.Fatalf("failed to marshal txDef in calTxHash: %s", err)
		return nil
	}
	allBytes := append(txDefBytes, uidBytes...)
	return calHash(allBytes)
}

func (c *Committer) executeBlock(msg common.CommitMsg) ([]*common.TxDefMsg, []common.TxDefMsg, error, common.Block, *vm.VM) {
	lastHeightBin := make([]byte, 8)
	if msg.Height == 0 {
		log.Fatalf("In CommitBlock: Height must greater than 0.")
	}
	binary.BigEndian.PutUint64(lastHeightBin, uint64(msg.Height-1))
	lastBlockBytes, err := c.BlockDB.Get(lastHeightBin, nil)
	if err != nil {
		log.Fatalf("failed to get last block[%d]: %s", msg.Height-1, err)
	}
	var lastBlock common.Block
	err = json.Unmarshal(lastBlockBytes, &lastBlock)
	if err != nil {
		log.Fatalf("failed to read last block: %s", err)
	}

	snapshot, err := c.StateDB.Get(lastHeightBin, nil)
	if err != nil {
		log.Fatalf("failed to open snapshot: %s", err)
	}
	engine := vm.NewVM(snapshot)

	successTxs := make([]common.TxDefMsg, 0)
	abortedTxs := make([]*common.TxDefMsg, 0)

	wmap := make(map[int][]byte)
	wg := sync.WaitGroup{}
	wg.Add(len(msg.Batch))
	mu := sync.Mutex{}
	for i, txDef := range msg.Batch {
		go func(localTxDef *common.TxDefMsg, localI int) {
			defer wg.Done()
			selfHash := calTxHash(localI, localTxDef)
			mu.Lock()
			defer mu.Unlock()

			if val, ok := wmap[localTxDef.IdxTo]; !ok || bytes.Compare(val, selfHash) > 0 {
				wmap[localTxDef.IdxTo] = selfHash
			}
		}(txDef, i)
	}
	wg.Wait() // 第一阶段预执行同步
	wg.Add(len(msg.Batch))
	for i, txDef := range msg.Batch {
		go func(localI int, localTxDef *common.TxDefMsg) {
			defer wg.Done()
			selfHash := calTxHash(localI, localTxDef)
			if c.checkConcurrent(wmap, localTxDef, selfHash, abortedTxs) {
				return
			}

			tx := utils.TxDefMsgToTransaction(localTxDef)
			innerErr := engine.ExecuteTransaction(tx)
			if innerErr != nil {
				log.Fatalf("failed to execute transaction: %s", err)
			}
			successTxs = append(successTxs, *localTxDef)
		}(i, txDef)
	}
	wg.Wait() // 第二阶段冲突处理同步
	return nil, successTxs, err, lastBlock, engine
}

func (c *Committer) checkConcurrent(wmap map[int][]byte, localTxDef *common.TxDefMsg, selfHash []byte, abortedTxs []*common.TxDefMsg) bool {
	c.wmapMutex.Lock()
	defer c.wmapMutex.Unlock()
	if val, ok := wmap[localTxDef.IdxTo]; ok && bytes.Compare(val, selfHash) < 0 {
		abortedTxs = append(abortedTxs, localTxDef)
		return true
	}
	if val, ok := wmap[localTxDef.IdxFrom]; ok && bytes.Compare(val, selfHash) < 0 {
		abortedTxs = append(abortedTxs, localTxDef)
		return true
	}
	return false
}
