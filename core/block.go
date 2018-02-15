package core

import (
	"time"
	"fmt"
	"crypto/sha512"
)

// interface to implement transactions
type Transaction interface {
	// execute using input data, and produce and output or error
	Execute(input interface{}) (interface{}, error)
}

// interface to implement block headers
type Header interface {
	Bytes() *Byte64
}

// interface for node information 
type NodeInfo interface{
	Id()	 string
}

// interface for block specification
type Block interface {
	Previous() Header
	Miner() NodeInfo
	Nonce() Header
	TD() time.Duration
	TXs() []Transaction
	Genesis() Header
	Hash() *Byte64
}

// a simple blockchain spec implementation
type SimpleBlock struct {
	previous Header
	miner NodeInfo
	hash *Byte64
	nonce Header
	genesis Header
	txs	[]Transaction
	td	  time.Duration
}

func (b *SimpleBlock) Previous() Header {
	return b.previous
}

func (b *SimpleBlock) Miner() NodeInfo {
	return b.miner
}

func (b *SimpleBlock) Nonce() Header {
	return b.nonce
}

func (b *SimpleBlock) TD() time.Duration {
	return b.td
}

func (b *SimpleBlock) Genesis() Header {
	return b.genesis
}

func (b *SimpleBlock) TXs() []Transaction {
	return b.txs
}

func (b *SimpleBlock) Hash() *Byte64 {
	return b.hash
}

type SimpleHeader struct {
	header *Byte64
}

func (h *SimpleHeader) Bytes() *Byte64 {
	return h.header
}

func NewSimpleHeader(header *Byte64) *SimpleHeader {
	return &SimpleHeader {
		header: header,
	}
}

func (b *SimpleBlock) ComputeHash() {
	data := make([]byte,0)
	data = append(data, b.previous.Bytes().Bytes()...)
	data = append(data, []byte(b.miner.Id())...)
	data = append(data, b.genesis.Bytes().Bytes()...)
	data = append(data, []byte(fmt.Sprintf("%d",b.td))...)
	hash := sha512.Sum512(data)
	b.hash = BytesToByte64(hash[:])
}

type SimpleNodeInfo struct {
	id string
}

func (n *SimpleNodeInfo) Id() string {
	return n.id
}

func NewSimpleNodeInfo(id string) *SimpleNodeInfo {
	return &SimpleNodeInfo{
		id: id,
	}
}

func NewSimpleBlock(previous *Byte64, genesis *Byte64, miner NodeInfo) *SimpleBlock {
	return &SimpleBlock{
		previous: NewSimpleHeader(previous),
		miner: miner,
		genesis: NewSimpleHeader(genesis),
		td: time.Duration(time.Now().UnixNano()),
		txs: make([]Transaction,0),
		hash: nil,
	}
}
