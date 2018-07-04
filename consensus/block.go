package consensus

import (
	"sync"
	"time"
	"crypto/sha512"
	"encoding/gob"
	"github.com/trust-net/go-trust-net/core"
	"github.com/trust-net/go-trust-net/core/trie"
	"github.com/trust-net/go-trust-net/common"
)

var (
	computeHashTimeoutSec = 10
)

type Block interface {
	ParentHash() *core.Byte64
	Miner() []byte
	Nonce() *core.Byte8
	Timestamp() *core.Byte8
	Delta() *core.Byte8
	Depth() *core.Byte8
	Weight() *core.Byte8
	Update(key, value []byte) bool
	Delete(key []byte) bool
	Lookup(key []byte) ([]byte, error)
	Uncles() []core.Byte64
	UncleMiners() [][]byte
	Transactions() []Transaction
	AddTransaction(tx *Transaction) error
	Hash() *core.Byte64
	Spec() BlockSpec
	// a deterministic numeric value for the block for ordering of competing blocks 
	Numeric() uint64
}

// these are the fields that actually go over the wire
type BlockSpec struct {
	PHASH core.Byte64
	MINER []byte
	STATE core.Byte64
	TXs []Transaction
	TS core.Byte8
	DELTA core.Byte8
	DEPTH core.Byte8
	WT core.Byte8
	UNCLEs []core.Byte64
	NONCE core.Byte8
}

func init() {
	gob.Register(&BlockSpec{})
	gob.Register(&block{})
}

type block struct {
	BlockSpec
	uncleMiners [][]byte
	pow PowApprover
	worldState trie.WorldState
	hash *core.Byte64
	isNetworkBlock bool
	variables map[string][]byte
	transactions map[core.Byte64]bool
	lock sync.RWMutex
}

func (b *block) ParentHash() *core.Byte64 {
	return &b.PHASH
}

func (b *block) Miner() []byte {
	return b.MINER
}

func (b *block) Nonce() *core.Byte8 {
	return &b.NONCE
}

func (b *block) Timestamp() *core.Byte8 {
	return &b.TS
}

func (b *block) Delta() *core.Byte8 {
	return &b.DELTA
}

func (b *block) Depth() *core.Byte8 {
	return &b.DEPTH
}

func (b *block) Weight() *core.Byte8 {
	return &b.WT
}

func (b *block) Update(key, value []byte) bool {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.variables[string(key)] = append(make([]byte, 0, len(value)), value...)
	return true
}

func (b *block) Delete(key []byte) bool {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.variables[string(key)] = nil
	return true
}

func (b *block) Lookup(key []byte) ([]byte, error) {
	b.lock.Lock()
	defer b.lock.Unlock()
	if value, ok := b.variables[string(key)]; ok {
		if value == nil {
			return nil, core.NewCoreError(ERR_KEY_NOT_FOUND, "key not found")
		} else {
			return value, nil
		}
	} else {
		value, err := b.worldState.Lookup(key)
		if err == nil {
			b.variables[string(key)] = append(make([]byte, 0, len(value)), value...)
		}
		return value, err
	}
}

func (b *block) Uncles() []core.Byte64 {
	return b.UNCLEs
}

func (b *block) UncleMiners() [][]byte {
	return b.uncleMiners
}

func (b *block) addUncle(uncle uncle) {
	b.UNCLEs = append(b.UNCLEs, *uncle.hash)
	b.uncleMiners = append(b.uncleMiners, uncle.miner)
	b.WT = *core.Uint64ToByte8(b.WT.Uint64()+1)
}

func (b *block) Transactions() []Transaction {
	return b.TXs
}

func (b *block) AddTransaction(tx *Transaction) error {
	b.lock.Lock()
	defer b.lock.Unlock()
	// first check if transaction does not already exists in parent's world state view
	if _, err := b.worldState.HasTransaction(tx.Id()); err == nil {
		return core.NewCoreError(ERR_DUPLICATE_TX, "duplicate transaction")
	}
	// now check if transaction was not already added to this block
	if _, found := b.transactions[*tx.Id()]; found {
		return core.NewCoreError(ERR_DUPLICATE_TX, "duplicate transaction")
	} else {
		b.transactions[*tx.Id()] = true
	}
	// accept transaction to the list 
	b.TXs = append(b.TXs, *tx)
	// not updating world state with transactions yet because don't have hash computed yet,
	// this will be done after hash computation, in the computeHash method
	return nil
}

func (b *block) Hash() *core.Byte64 {
	return b.hash
}

func (b *block) persistState() error {
	// we don't want to cleanup the original state, since it is parent block's world state
	skippedParentState := false
	// update variables
	for key, value := range b.variables {
		oldHash := b.worldState.Hash()
		var newHash core.Byte64
		if value == nil {
			newHash = b.worldState.Delete([]byte(key))
		} else {
			newHash = b.worldState.Update([]byte(key), value)
		}
		// if hash did not change, skip
		if  newHash == oldHash {
			continue
		}
		// cleanup old transient hash (non parent hash)
		if skippedParentState {
			if err := b.worldState.Cleanup(oldHash); err != nil {
//				log.AppLogger().Error("Failed to cleanup state: %s", err)
			}
		} else {
			skippedParentState = true
		}
	}
	// transaction will be registered after compute hash, since need block hash
	return nil
}

// block hash = SHA512(parent_hash + author_node + timestamp + state + depth + transactions... + weight + uncles... + nonce)
func (b *block) computeHash() *core.Byte64 {
	b.lock.Lock()
	defer b.lock.Unlock()
	// parent hash +
	// miner ID +
	// block timestampt +
	// time since parent block +
	// world state fingerprint +
	// block's depth from genesis +
	// transactions... +
	// block's weight +
	// block's uncle's hash +
	// nonce
	data := make([]byte,0, 64+64+8+64+8+len(b.TXs)*(8+64+1)+8+len(b.UNCLEs)*64)
	data = append(data, b.PHASH.Bytes()...)
	data = append(data, b.MINER...)
	data = append(data, b.TS.Bytes()...)
	data = append(data, b.DELTA.Bytes()...)
	var statePtr *core.Byte64
	if b.worldState != nil {
		// persist the world state
		if b.persistState() != nil {
			// return error value
			return nil
		}
		// if its not network block, then update state
		if !b.isNetworkBlock {
			b.STATE = b.worldState.Hash()
		}
		state := b.worldState.Hash()
		statePtr = &state
	} else {
		statePtr = &b.STATE
	}
	data = append(data, statePtr.Bytes()...)
	for _, tx := range b.TXs {
		data = append(data, tx.Bytes()...)
	}
	data = append(data, b.WT.Bytes()...)
	for _, uncle := range b.UNCLEs {
		data = append(data, uncle.Bytes()...)
	}
	dataWithNonce := make([]byte, 0, len(data)+8)
	nonce := b.NONCE.Uint64()
	var hash [sha512.Size]byte
	isPoWDone := false

	// run the tight loop below in a timeout thread
	err := common.RunTimeBoundSec(computeHashTimeoutSec, func() error {
			for !isPoWDone {
				// TODO: run the PoW
				b.NONCE = *core.Uint64ToByte8(nonce)
				nonce++
				dataWithNonce = append(data, b.NONCE.Bytes()...)
				hash = sha512.Sum512(dataWithNonce)
				// check PoW validation
				if b.pow != nil {
					isPoWDone = b.pow(hash[:], b.TS.Uint64(), b.DELTA.Uint64())
				} else {
					isPoWDone = true
				}
				
				// if a network block, then 1st hash MUST be correct
				if !isPoWDone && b.isNetworkBlock {
					// return an error
					return core.NewCoreError(ERR_HASH_INCORRECT, "invalid hash on network block")
				}
			}
			return nil
		}, core.NewCoreError(ERR_HASH_TIMEOUT, "compute hash timeout"))
	if err != nil {
		return nil
	}
	b.hash = core.BytesToByte64(hash[:])
	return b.hash
}

func (b *block) registerTransactions() error {
	if b.worldState != nil {
		for _, tx := range b.TXs {
			if err := b.worldState.RegisterTransaction(tx.Id(), b.hash); err != nil {
				return err
			}
		}
	}
	return nil	
}

// create a copy of block
func (b *block) clone(state trie.WorldState) *block {
	clone := &block{
		BlockSpec: BlockSpec {
			PHASH: b.PHASH,
			MINER: append([]byte{}, b.MINER...),
			STATE: state.Hash(),
			TXs: nil,
			TS: b.TS,
			DELTA: b.DELTA,
			DEPTH: b.DEPTH,
			WT: b.WT,
			UNCLEs: nil,
			NONCE: b.NONCE,
		},
		hash: b.hash,
		worldState: state,
		isNetworkBlock: false,
		variables: make(map[string][]byte),
		transactions: make(map[core.Byte64]bool),
	}
	return clone
}

// create a copy of block sendable on wire
func (b *block) Spec() BlockSpec {
	spec := BlockSpec{
		PHASH: b.PHASH,
		MINER: append([]byte{}, b.MINER...),
		TXs: make([]Transaction,len(b.TXs)),
		TS: b.TS,
		DELTA: b.DELTA,
		DEPTH: b.DEPTH,
		WT: b.WT,
		UNCLEs: make([]core.Byte64,len(b.UNCLEs)),
		NONCE: b.NONCE,
	}
	if b.worldState != nil {
		spec.STATE = b.worldState.Hash()
	} else {
		spec.STATE = b.STATE
	}
	for i, tx := range b.TXs {
		spec.TXs[i] = tx
	}
	for i, uncle := range b.UNCLEs {
		spec.UNCLEs[i] = uncle
	}
	return spec
}

// a deterministic numeric value for the block for ordering of competing blocks 
func (b *block) Numeric() uint64 {
	num := uint64(0)
	if b.hash == nil {
		num -= 1
		return num
	}
	for _, b := range b.hash.Bytes() {
		num += uint64(b)
	}
	return num
}

// private method, can only be invoked by DAG implementation, so can be initialized correctly
func newBlock(previous *core.Byte64, weight uint64, depth uint64, ts, pTs uint64, miner []byte, state trie.WorldState) *block {
	if ts == 0 {
		ts = uint64(time.Now().UnixNano())
	}
	b := &block{
		BlockSpec: BlockSpec {
			PHASH: *previous,
			MINER: miner,
			TXs: make([]Transaction,0,1),
			TS: *core.Uint64ToByte8(ts),
			DELTA: *core.Uint64ToByte8(ts - pTs),
			DEPTH: *core.Uint64ToByte8(depth),
			WT: *core.Uint64ToByte8(weight),
			UNCLEs: make([]core.Byte64, 0),
			NONCE: *core.BytesToByte8(nil),
		},
		worldState: state,
		hash: nil,
		variables: make(map[string][]byte),
		transactions: make(map[core.Byte64]bool),
	}
	if state != nil {
		b.STATE = state.Hash()
	}
	return b
}

func serializeBlock(b Block) ([]byte, error) {
	if b == nil {
		return nil, core.NewCoreError(ERR_INVALID_ARG, "nil block")
	}
	block, ok := b.(*block)
	if !ok {
		return nil, core.NewCoreError(ERR_TYPE_INCORRECT, "incorrect type")
	}
	if block.STATE == *core.BytesToByte64(nil) || (block.worldState != nil && block.worldState.Hash() != block.STATE) {
		return nil, core.NewCoreError(ERR_STATE_INCORRECT, "block state incorrect")
	}
	if block.hash == nil {
		return nil, core.NewCoreError(ERR_BLOCK_UNHASHED, "block not hashed")
	}
	return common.Serialize(b)
}

// private method, can only be invoked by DAG implementation, so that world state can be added after deserialization
// here we only want to deserialize wire protocol data into block instance, then after this DAG implementation
// will use a world state rebased to parent's state trie, and then pass it on to application to run
// the transactions and value changes as appropriate
func deSerializeBlock(data []byte) (*block, error) {
	var b block
	if err := common.Deserialize(data, &b); err != nil {
		return nil, err
	}
	b.isNetworkBlock = true
	b.variables = make(map[string][]byte)
	b.transactions = make(map[core.Byte64]bool)

	// Q: when, where, who to update world state with this block's value changes?
	// A: application will validate transactions, at which time world state will be updated with values
	//    and then submit the network block for acceptance, at which time canonical chain will start pointing
	//    to world state view of this block (if accepted)
	return &b, nil
}