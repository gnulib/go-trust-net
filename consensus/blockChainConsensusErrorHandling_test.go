package consensus

import (
    "testing"
    "time"
	"github.com/trust-net/go-trust-net/core"
	"github.com/trust-net/go-trust-net/db"
	"github.com/trust-net/go-trust-net/log"
)

// following is an implementation of DB interface that can be mocked out to send specific error responses
type errorDb struct {
	step int
	values []error
}
func (db *errorDb) Put(key []byte, value []byte) error {
	db.step++
	return db.values[db.step-1]
}

func (db *errorDb) Get(key []byte) ([]byte, error) {
	db.step++
	return nil, db.values[db.step-1]
}

func (db *errorDb) Has(key []byte) (bool, error) {
	db.step++
	return false, db.values[db.step-1]
}

func (db *errorDb) Delete(key []byte) error {
	db.step++
	return db.values[db.step-1]
}

func (db *errorDb) Close() error{
	db.step++
	return db.values[db.step-1]
}

type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}

func TestErrorNewBlockChainConsensusDagTipSaveError(t *testing.T) {
	log.SetLogLevel(log.NONE)
	db := &errorDb {
		values: []error{
			nil, // trie op
			nil, // trie op
			&testError{"get dag tip error"},
			&testError{"put dag tip error"},
		},
	}
	// verify that blockchain reports error when cannot save dag tip 
	_, err := NewBlockChainConsensus(genesisHash, genesisTime, testNode, db)
	if err == nil || err.Error() != "put dag tip error" {
		t.Errorf("failed to report error when cannot save DAG tip: %s", err)
		return
	}
}

func TestErrorNewBlockChainConsensusGenesisBlockSaveError(t *testing.T) {
	log.SetLogLevel(log.NONE)
	db := &errorDb {
		values: []error{
			nil, // trie op
			nil, // trie op
			&testError{"get dag tip error"},
			nil, // put dag tip
			&testError{"put genesis block error"},
		},
	}
	// verify that blockchain reports error when cannot save dag tip 
	_, err := NewBlockChainConsensus(genesisHash, genesisTime, testNode, db)
	if err == nil || err.Error() != "put genesis block error" {
		t.Errorf("failed to report error when cannot save genesis block: %s", err)
		return
	}
}


func TestErrorNewBlockChainConsensusGenesisChainNodeSaveError(t *testing.T) {
	log.SetLogLevel(log.NONE)
	db := &errorDb {
		values: []error{
			nil, // trie op
			nil, // trie op
			&testError{"get dag tip error"},
			nil, // put dag tip
			nil, // put genesis block
			&testError{"put genesis chain node error"},
		},
	}
	// verify that blockchain reports error when cannot save dag tip 
	_, err := NewBlockChainConsensus(genesisHash, genesisTime, testNode, db)
	if err == nil || err.Error() != "put genesis chain node error" {
		t.Errorf("failed to report error when cannot save genesis block: %s", err)
		return
	}
}

func TestErrorNewBlockChainConsensusGetTipBlockError(t *testing.T) {
	log.SetLogLevel(log.NONE)
	db := &errorDb {
		values: []error{
			nil, // trie op
			nil, // trie op
			nil, // get dag tip
			&testError{"get dag block error"},
		},
	}
	// verify that blockchain reports error when cannot save dag tip 
	_, err := NewBlockChainConsensus(genesisHash, genesisTime, testNode, db)
	if err == nil || err.Error() != "get dag block error" {
		t.Errorf("failed to report error when cannot get DAG block: %s", err)
		return
	}
}

func TestErrorNewBlockChainConsensusGetBlockDbError(t *testing.T) {
	log.SetLogLevel(log.NONE)
	db, _ := db.NewDatabaseInMem()
	// verify that blockchain reports error when cannot get block 
	c, err := NewBlockChainConsensus(genesisHash, genesisTime, testNode, db)
	// override chain db to a mock and return error
	c.db = &errorDb {
		values: []error{
			&testError{"get block error"},
		},
	}
	_, err = c.getBlock(core.BytesToByte64(nil))
	if err == nil || err.Error() != "get block error" {
		t.Errorf("failed to report error when cannot get block from db: %s", err)
		return
	}
}

func TestErrorNewBlockChainConsensusGetBlockDeSerializeError(t *testing.T) {
	log.SetLogLevel(log.NONE)
	db, _ := db.NewDatabaseInMem()
	// verify that blockchain reports error when cannot deserialize block 
	c, err := NewBlockChainConsensus(genesisHash, genesisTime, testNode, db)
	// override chain db to a mock and return error
	c.db = &errorDb {
		values: []error{
			nil, // get block with nil data
		},
	}
	_, err = c.getBlock(core.BytesToByte64(nil))
	if err == nil || err.Error() != "EOF" {
		t.Errorf("failed to report error when cannot deserialize block from db: %s", err)
		return
	}
}

func TestErrorNewBlockChainConsensusPutBlockError(t *testing.T) {
	log.SetLogLevel(log.NONE)
	db, _ := db.NewDatabaseInMem()
	// verify that blockchain reports error when cannot save block 
	c, err := NewBlockChainConsensus(genesisHash, genesisTime, testNode, db)
	// override chain db to a mock and return error
	block := newBlock(c.Tip().Hash(), c.Tip().Weight().Uint64() + 1, c.Tip().Depth().Uint64() + 1, uint64(time.Now().UnixNano()), c.minerId, c.state)
	err = c.putBlock(block)
	if err == nil || err.(*core.CoreError).Code() != ERR_BLOCK_UNHASHED {
		t.Errorf("failed to report error when cannot put block into db: %s", err)
		return
	}
}

func TestErrorNewBlockChainConsensusGetChainNodeDbError(t *testing.T) {
	log.SetLogLevel(log.NONE)
	db, _ := db.NewDatabaseInMem()
	// verify that blockchain reports error when cannot save dag tip 
	c, err := NewBlockChainConsensus(genesisHash, genesisTime, testNode, db)
	// override chain db to a mock and return error
	c.db = &errorDb {
		values: []error{
			&testError{"get chain node error"},
		},
	}
	_, err = c.getChainNode(core.BytesToByte64(nil))
	if err == nil || err.Error() != "get chain node error" {
		t.Errorf("failed to report error when cannot get chain node from db: %s", err)
		return
	}
}

func TestErrorNewBlockChainConsensusGetChainNodeDeSerializeError(t *testing.T) {
	log.SetLogLevel(log.NONE)
	db, _ := db.NewDatabaseInMem()
	// verify that blockchain reports error when cannot save dag tip 
	c, err := NewBlockChainConsensus(genesisHash, genesisTime, testNode, db)
	// override chain db to a mock and return error
	c.db = &errorDb {
		values: []error{
			nil, // get block with nil data
		},
	}
	_, err = c.getChainNode(core.BytesToByte64(nil))
	if err == nil || err.Error() != "EOF" {
		t.Errorf("failed to report error when cannot deserialize chain node from db: %s", err)
		return
	}
}

func TestErrorDeserializeNetworkBlockDeSerializeError(t *testing.T) {
	log.SetLogLevel(log.NONE)
	db, _ := db.NewDatabaseInMem()
	consensus, err := NewBlockChainConsensus(genesisHash, genesisTime, testNode, db)
	if err != nil || consensus == nil {
		t.Errorf("failed to get blockchain consensus instance: %s", err)
		return
	}
	_, err = consensus.DeserializeNetworkBlock(nil)
	if err == nil || err.Error() != "EOF" {
		t.Errorf("failed to report error when cannot deserialize network block: %s", err)
		return
	}
}

func TestErrorDeserializeNetworkBlockNoParent(t *testing.T) {
	log.SetLogLevel(log.NONE)
	db, _ := db.NewDatabaseInMem()
	c, err := NewBlockChainConsensus(genesisHash, genesisTime, testNode, db)
	if err != nil || c == nil {
		t.Errorf("failed to get blockchain consensus instance: %s", err)
		return
	}
	// build a new block that does not have its parent in the chain
	child := newBlock(core.BytesToByte64([]byte("some random parent")), 100, 100, uint64(time.Now().UnixNano()), testNode, c.state)
	child.computeHash()
	data,_ := serializeBlock(child)
	if _, err = c.DeserializeNetworkBlock(data); err == nil {
		t.Errorf("failed to detect orphan network block")
	}
}

func TestErrorDeserializeNetworkBlockIncorrectDepth(t *testing.T) {
	log.SetLogLevel(log.NONE)
	db, _ := db.NewDatabaseInMem()
	c, err := NewBlockChainConsensus(genesisHash, genesisTime, testNode, db)
	if err != nil || c == nil {
		t.Errorf("failed to get blockchain consensus instance: %s", err)
		return
	}
	// build a new block as current tip's child, but incorrect depth
	child := newBlock(c.Tip().Hash(), c.Tip().Weight().Uint64() + 1, c.Tip().Depth().Uint64() + 100, uint64(time.Now().UnixNano()), c.minerId, c.state)
	child.computeHash()
	data,_ := serializeBlock(child)
	if _, err = c.DeserializeNetworkBlock(data); err == nil {
		t.Errorf("failed to detect incorrect depth on network block")
	}
}

func TestErrorDeserializeNetworkBlockIncorrectWeight(t *testing.T) {
	log.SetLogLevel(log.NONE)
	db, _ := db.NewDatabaseInMem()
	c, err := NewBlockChainConsensus(genesisHash, genesisTime, testNode, db)
	if err != nil || c == nil {
		t.Errorf("failed to get blockchain consensus instance: %s", err)
		return
	}
	// build a new block as current tip's child, but incorrect weight
	child := newBlock(c.Tip().Hash(), c.Tip().Weight().Uint64(), c.Tip().Depth().Uint64() + 1, uint64(time.Now().UnixNano()), c.minerId, c.state)
	child.computeHash()
	data,_ := serializeBlock(child)
	if _, err = c.DeserializeNetworkBlock(data); err == nil {
		t.Errorf("failed to detect incorrect weight on network block")
	}
}


func TestErrorDeserializeNetworkBlockIncorrectUncle(t *testing.T) {
	log.SetLogLevel(log.NONE)
	db, _ := db.NewDatabaseInMem()
	c, err := NewBlockChainConsensus(genesisHash, genesisTime, testNode, db)
	if err != nil || c == nil {
		t.Errorf("failed to get blockchain consensus instance: %s", err)
		return
	}
	// build a new block as current tip's child, but incorrect uncle
	child := newBlock(c.Tip().Hash(), c.Tip().Weight().Uint64()+1+1, c.Tip().Depth().Uint64()+1, uint64(time.Now().UnixNano()), c.minerId, c.state)
	child.UNCLEs = append(child.UNCLEs, *core.BytesToByte64([]byte("invalid uncle")))
	child.computeHash()
	data,_ := serializeBlock(child)
	if _, err = c.DeserializeNetworkBlock(data); err == nil {
		t.Errorf("failed to detect incorrect uncle on network block")
	}
}
