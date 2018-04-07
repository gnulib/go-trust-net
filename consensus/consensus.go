package consensus

import (
	"github.com/trust-net/go-trust-net/core"
	"github.com/ethereum/go-ethereum/p2p"
)

// a callback provided by application to handle results of block mining request
// block instance is provided (to update nodeset with hash), if successful, or error is provided if failed/aborted
type MiningResultHandler func(block Block, err error)
// a callback provided by application to approve PoW
// (arguments include block's timestamp and delta time since parent block,
// so that application can implement variable PoW schemes based on time when
// block was generated and time since its parent)
type PowApprover	func(hash []byte, ts, delta uint64) bool

// a consensus platform interface
type Consensus interface {
	// get a new "candidate" block, initialized with a copy of world state
	// from request time tip of the canonical chain
	NewCandidateBlock() Block
	// submit a "filled" block for mining (executes as a goroutine)
	// it will mine the block and update canonical chain or abort if a new network block
	// is received with same or higher weight, the callback MiningResultHandler will be called
	// with serialized data for the block that can be  sent over the wire to peers,
	// or error if mining failed/aborted
	MineCandidateBlock(b Block, cb MiningResultHandler)
	// a PoW variant of the MineCandidateBlock method
	MineCandidateBlockPoW(b Block, apprvr PowApprover, cb MiningResultHandler)
	// query status of a transaction (its block details) in the canonical chain
	TransactionStatus(txId *core.Byte64) (Block, error)
	// deserialize data into network block, and will initialize the block with current canonical parent's
	// world state root (application is responsible to run the transactions from block, and update
	// world state appropriately)
	DeserializeNetworkBlock(data []byte) (Block, error)
	// an equivalend implementation, to decode devP2P network block message
	DecodeNetworkBlock(msg p2p.Msg) (Block, error)
	// an equivalend implementation, to decode network block spec message
	DecodeNetworkBlockSpec(spec BlockSpec) (Block, error)
	// submit a "processed" network block, will be added to DAG appropriately
	// (i.e. either extend canonical chain, or add as an uncle block)
	// block's computed world state should match STATE of the deSerialized block,
	AcceptNetworkBlock(b Block) error
    // a copy of best block in current cannonical chain, used by protocol manager for handshake
    BestBlock() Block
    // genesis Hash
    Genesis() *core.Byte64
    // a copy of block with specified hash, or error if not found
    Block(hash *core.Byte64) (Block, error)
    // ordered list of descendents from specific parent, on the current canonical chain
    Descendents(parent *core.Byte64, max int) ([]Block, error)
    // the ancestor at max distance from specified child
    Ancestor(child *core.Byte64, max int) (Block, error)
}