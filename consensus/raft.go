package consensus

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"neochain/commit"
	"neochain/common"
	"neochain/utils"
	"strings"
	"sync"
	"time"

	pb "github.com/Jille/raft-grpc-example/proto"
	"github.com/Jille/raft-grpc-leader-rpc/rafterrors"
	"github.com/hashicorp/raft"
)

const BLOCK_SIZE = 128

// Raft keeps track of the three longest queue it ever saw.
type Raft struct {
	mtx      sync.RWMutex
	queue    []*common.TxDefMsg
	epoch    int
	commiter *commit.Committer
}

var _ raft.FSM = &Raft{}

func NewRaftEngine(commiter *commit.Committer) *Raft {
	return &Raft{
		queue:    make([]*common.TxDefMsg, 0),
		epoch:    1,
		commiter: commiter,
	}
}

// Apply 最终效果只是增加一个word
func (f *Raft) Apply(l *raft.Log) interface{} {

	f.mtx.Lock()
	defer f.mtx.Unlock()

	w := string(l.Data)
	msg, err := utils.JsonToTxDefMsg(w)
	if err != nil {
		//log.Fatalf("TxDefMsg deserialization err: %v", err)
		return fmt.Errorf("TxDefMsg deserialization err: %v", err)
	}
	log.Printf("Receive a msg: %v, len(queue)=%d, epoch=%d", msg, len(f.queue), f.epoch)
	f.queue = append(f.queue, msg)
	for len(f.queue) >= f.epoch*BLOCK_SIZE {
		consensusComplete := time.Now()
		log.Printf("performance statistic: consensus[e][%d]: %v", f.epoch, consensusComplete)

		indexFrom := (f.epoch - 1) * BLOCK_SIZE
		go func(localEpoch int) {
			abortedTxs := f.commiter.CommitBlock(common.CommitMsg{
				Batch:  f.queue[indexFrom : indexFrom+BLOCK_SIZE],
				Height: localEpoch,
			})

			f.queue = append(abortedTxs, f.queue...)
		}(f.epoch)

		f.epoch++
		consensusStart := time.Now()
		log.Printf("performance statistic: consensus[s][%d]: %v", f.epoch, consensusStart)
	}
	return nil
}

func (f *Raft) Snapshot() (raft.FSMSnapshot, error) {
	// Make sure that any future calls to f.Apply() don't change the snapshot.
	return &snapshot{pool: clonePool(f.queue)}, nil
}

func clonePool(queue []*common.TxDefMsg) []*common.TxDefMsg {
	copyQ := make([]*common.TxDefMsg, len(queue))
	copy(copyQ, queue)
	return copyQ
}

func (f *Raft) Restore(r io.ReadCloser) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	words := strings.Split(string(b), "\n")
	f.queue = make([]*common.TxDefMsg, len(words))
	for i, word := range words {
		json, innerErr := utils.JsonToTxDefMsg(word)
		if innerErr != nil {
			return fmt.Errorf("TxDefMsg deserialization err: %v", innerErr)
		}
		f.queue[i] = json
	}
	return nil
}

type snapshot struct {
	pool []*common.TxDefMsg
}

func (s *snapshot) Persist(sink raft.SnapshotSink) error {
	jsonMsgs := make([]string, len(s.pool))
	for i, m := range s.pool {
		jsonStr, err := utils.TxDefMsgToJson(m)
		if err != nil {
			return err
		}
		jsonMsgs[i] = jsonStr
	}
	_, err := sink.Write([]byte(strings.Join(jsonMsgs, "\n")))
	if err != nil {
		innerErr := sink.Cancel()
		if innerErr != nil {
			return fmt.Errorf("sink.Cancel() err: %v", innerErr)
		}
		return fmt.Errorf("sink.Write(): %v", err)
	}
	return sink.Close()
}

func (s *snapshot) Release() {
}

type RpcInterface struct {
	WordTracker *Raft
	Raft        *raft.Raft
}

func (r RpcInterface) AddWord(ctx context.Context, req *pb.AddWordRequest) (*pb.AddWordResponse, error) {
	//log.Printf("performance statistic: Start time: %v AddWord req: %v", start, req)

	//time.Sleep(time.Millisecond * time.Duration(300*rand.Float32())) // consensus is too fast

	f := r.Raft.Apply([]byte(req.GetWord()), time.Second)
	if err := f.Error(); err != nil {
		return nil, rafterrors.MarkRetriable(err)
	}
	return &pb.AddWordResponse{
		CommitIndex: f.Index(),
	}, nil
}

func (r RpcInterface) GetWords(ctx context.Context, req *pb.GetWordsRequest) (*pb.GetWordsResponse, error) {
	r.WordTracker.mtx.RLock()
	defer r.WordTracker.mtx.RUnlock()
	poolJson := make([]string, len(r.WordTracker.queue))
	for i, m := range r.WordTracker.queue {
		msgJson, err := utils.TxDefMsgToJson(m)
		if err != nil {
			return nil, fmt.Errorf("TxDefMsg serialization err: %v", err)
		}
		poolJson[i] = msgJson
	}
	return &pb.GetWordsResponse{
		BestWords:   poolJson,
		ReadAtIndex: r.Raft.AppliedIndex(),
	}, nil
}
