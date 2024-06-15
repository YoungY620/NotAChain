package commit

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"log"
	"neochain/common"
	"neochain/vm"
	"sync"
	"time"
)

type Committer struct {
	BlockDB *leveldb.DB
	StateDB *leveldb.DB

	mutex sync.Mutex
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

func (c *Committer) CommitBlock(msg common.CommitMsg) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	log.Printf("performance statistic: exe[s][%d]: %v", msg.Height, time.Now().UnixNano())
	err, lastBlock, engine := c.executeBlock(msg)
	if err != nil {
		log.Fatalf("failed to execute block: %s", err)
	}
	log.Printf("performance statistic: exe[e][%d]: %v", msg.Height, time.Now().UnixNano())

	log.Printf("performance statistic: commit[s][%d]: %v", msg.Height, time.Now().UnixNano())
	block := &common.Block{
		Header: common.BlockHeader{
			Height:        msg.Height,
			PrevBlockHash: lastBlock.Header.BlockHash,
			BlockHash:     "",
		},
		Txs: msg.Batch,
	}
	c.storeBlock(msg.Height, block, engine.Context.Memory.Cell)
	log.Printf("performance statistic: commit[e][%d]: %v", msg.Height, time.Now().UnixNano())
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
	hash := sha256.New()
	hash.Write(blockBytes)
	block.Header.BlockHash = hex.EncodeToString(hash.Sum(nil))
	return blockBytes
}

func (c *Committer) executeBlock(msg common.CommitMsg) (error, common.Block, *vm.VM) {
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

	for _, tx := range msg.Batch {
		innerErr := engine.ExecuteTransaction(&tx)
		if innerErr != nil {
			log.Fatalf("failed to execute transaction: %s", err)
		}
	}
	return err, lastBlock, engine
}
