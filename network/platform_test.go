package network

import (
    "testing"
    "time"
	"github.com/trust-net/go-trust-net/db"
	"github.com/trust-net/go-trust-net/core"
	"github.com/trust-net/go-trust-net/consensus"
	"github.com/trust-net/go-trust-net/log"
	"github.com/ethereum/go-ethereum/p2p"
)

func TestNewPlatformManagerNullArgs(t *testing.T) {
	log.SetLogLevel(log.NONE)
	defer log.SetLogLevel(log.NONE)
	db, _ := db.NewDatabaseInMem()
	if _, err := NewPlatformManager(nil, &ServiceConfig{}, db); err == nil {
		t.Errorf("did not detect nil app config")
	}
	if _, err := NewPlatformManager(&AppConfig{}, nil, db); err == nil {
		t.Errorf("did not detect nil service config")
	}
	if _, err := NewPlatformManager(&AppConfig{}, &ServiceConfig{}, nil); err == nil {
		t.Errorf("did not detect nil app DB")
	}	
}

func TestNewPlatformManagerGoodArgs(t *testing.T) {
	log.SetLogLevel(log.NONE)
	defer log.SetLogLevel(log.NONE)
	db, _ := db.NewDatabaseInMem()
	conf := testNetworkConfig(nil, nil, nil)
	if mgr, err := NewPlatformManager(&conf.AppConfig, &conf.ServiceConfig, db); err != nil {
		t.Errorf("Failed to create platform manager: %s", err)
	} else {
		if mgr.PeerCount() != 0 {
			t.Errorf("did not expect any peers from new platform instance")
		}
	}
}

func TestValidateAndAddIncorrectNetwork(t *testing.T) {
	log.SetLogLevel(log.NONE)
	defer log.SetLogLevel(log.NONE)
	db, _ := db.NewDatabaseInMem()
	conf := testNetworkConfig(nil, nil, nil)
	if mgr, err := NewPlatformManager(&conf.AppConfig, &conf.ServiceConfig, db); err != nil {
		t.Errorf("Failed to create platform manager: %s", err)
	} else {
		// create a peer handshake message with network ID mismatch
		handshake := testPeerHandshakeMsg(mgr.config)
		handshake.NetworkId = *core.BytesToByte16([]byte("an invalid network id"))
		peer := &mockNode{}
		if err := mgr.validateAndAdd(handshake, peer); err == nil {
			t.Errorf("Failed to detect network mistmatch in peer handshake")
		}
		if mgr.PeerCount() != 0 {
			t.Errorf("did not expect any peers for incorrect handshake")
		}
	}
}

func TestValidateAndAddIncorrectGenesis(t *testing.T) {
	log.SetLogLevel(log.NONE)
	defer log.SetLogLevel(log.NONE)
	db, _ := db.NewDatabaseInMem()
	conf := testNetworkConfig(nil, nil, nil)
	if mgr, err := NewPlatformManager(&conf.AppConfig, &conf.ServiceConfig, db); err != nil {
		t.Errorf("Failed to create platform manager: %s", err)
	} else {
		// create a peer handshake message with network ID mismatch
		handshake := testPeerHandshakeMsg(mgr.config)
		handshake.Genesis = *core.BytesToByte64([]byte("an invalid genesis"))
		peer := &mockNode{}
		if err := mgr.validateAndAdd(handshake, peer); err == nil {
			t.Errorf("Failed to detect genesis mistmatch in peer handshake")
		}
		if mgr.PeerCount() != 0 {
			t.Errorf("did not expect any peers for incorrect handshake")
		}
	}
}

func TestValidateAndAddValidatorError(t *testing.T) {
	log.SetLogLevel(log.NONE)
	defer log.SetLogLevel(log.NONE)
	db, _ := db.NewDatabaseInMem()
	// create a network config with validator function that will reject peer connection
	conf := testNetworkConfig(nil, nil, func(config *AppConfig) error {
			return core.NewCoreError(0x11, "test error")
	})
	if mgr, err := NewPlatformManager(&conf.AppConfig, &conf.ServiceConfig, db); err != nil {
		t.Errorf("Failed to create platform manager: %s", err)
	} else {
		// create a peer handshake message with network ID mismatch
		handshake := testPeerHandshakeMsg(mgr.config)
		peer := &mockNode{
			peerNode: peerNode{
				peer: p2p.NewPeer([64]byte(*core.BytesToByte64([]byte("a random peer node ID"))), "", nil),
			},
		}
		if err := mgr.validateAndAdd(handshake, peer); err == nil || err.(*core.CoreError).Code() != 0x11 || err.Error() != "test error" {
			t.Errorf("Failed to detect application validation error in peer handshake: %s", err)
		}
		if mgr.PeerCount() != 0 {
			t.Errorf("did not expect any peers for incorrect handshake")
		}
	}
}

func TestValidateAndAddPeerSetUpdate(t *testing.T) {
	log.SetLogLevel(log.NONE)
	defer log.SetLogLevel(log.NONE)
	db, _ := db.NewDatabaseInMem()
	// create a network config with validator function that accept peer connection
	conf := testNetworkConfig(nil, nil, func(config *AppConfig) error {
			return nil
	})
	if mgr, err := NewPlatformManager(&conf.AppConfig, &conf.ServiceConfig, db); err != nil {
		t.Errorf("Failed to create platform manager: %s", err)
	} else {
		// create a peer handshake message with network ID mismatch
		handshake := testPeerHandshakeMsg(mgr.config)
		peer := &mockNode{
			peerNode: peerNode{
				peer: p2p.NewPeer([64]byte(*core.BytesToByte64([]byte("a random peer node ID"))), "", nil),
			},
		}
		if err := mgr.validateAndAdd(handshake, peer); err != nil {
			t.Errorf("Failed to validate peer handshake: %s", err)
		}
		if mgr.PeerCount() != 1 {
			t.Errorf("did not update peer count")
		}
		if mgr.peerDb.PeerNodeForId(peer.Id()) != peer {
			t.Errorf("did not find peer in peer set DB")
		}
	}
}

func TestUnregisterPeer(t *testing.T) {
	log.SetLogLevel(log.NONE)
	defer log.SetLogLevel(log.NONE)
	db, _ := db.NewDatabaseInMem()
	// create a network config with validator function that accept peer connection
	conf := testNetworkConfig(nil, nil, func(config *AppConfig) error {
			return nil
	})
	if mgr, err := NewPlatformManager(&conf.AppConfig, &conf.ServiceConfig, db); err != nil {
		t.Errorf("Failed to create platform manager: %s", err)
	} else {
		// create a peer handshake message with network ID mismatch
		handshake := testPeerHandshakeMsg(mgr.config)
		peer := &mockNode{
			peerNode: peerNode{
				peer: p2p.NewPeer([64]byte(*core.BytesToByte64([]byte("a random peer node ID"))), "", nil),
			},
		}
		if err := mgr.validateAndAdd(handshake, peer); err != nil {
			t.Errorf("Failed to validate peer handshake: %s", err)
		}
		// now disconnect peer
		mgr.UnregisterPeer(peer)
		if mgr.PeerCount() != 0 {
			t.Errorf("did not decrement peer count")
		}
		if mgr.peerDb.PeerNodeForId(peer.Id()) != nil {
			t.Errorf("remove peer from peer set DB")
		}
	}
}

func TestHandshakeMsgSendErr(t *testing.T) {
	log.SetLogLevel(log.NONE)
	defer log.SetLogLevel(log.NONE)
	db, _ := db.NewDatabaseInMem()
	// create a network config with validator function that accept peer connection
	conf := testNetworkConfig(nil, nil, func(config *AppConfig) error {
			return nil
	})
	if mgr, err := NewPlatformManager(&conf.AppConfig, &conf.ServiceConfig, db); err != nil {
		t.Errorf("Failed to create platform manager: %s", err)
	} else {
		// create a peer handshake message with network ID mismatch
		handshake := testPeerHandshakeMsg(mgr.config)
		peer := &mockNode{
			peerNode: peerNode{
				peer: p2p.NewPeer([64]byte(*core.BytesToByte64([]byte("a random peer node ID"))), "", nil),
			},
			testMocks: testMocks{
				sendErr: core.NewCoreError(0x2100000, "test send error"),
			},
		}
		if err := mgr.Handshake(handshake, peer); err == nil || err.(*core.CoreError).Code() != ErrorHandshakeFailed {
			t.Errorf("Failed to detect handshake msg send error: %s", err)
			return
		}
	}
}

func TestHandshakeMsgReadErr(t *testing.T) {
	log.SetLogLevel(log.NONE)
	defer log.SetLogLevel(log.NONE)
	db, _ := db.NewDatabaseInMem()
	// create a network config with validator function that accept peer connection
	conf := testNetworkConfig(nil, nil, func(config *AppConfig) error {
			return nil
	})
	if mgr, err := NewPlatformManager(&conf.AppConfig, &conf.ServiceConfig, db); err != nil {
		t.Errorf("Failed to create platform manager: %s", err)
	} else {
		// create a peer handshake message with network ID mismatch
		handshake := testPeerHandshakeMsg(mgr.config)
		peer := &mockNode{
			peerNode: peerNode{
				peer: p2p.NewPeer([64]byte(*core.BytesToByte64([]byte("a random peer node ID"))), "", nil),
			},
			testMocks: testMocks{
				readErr: core.NewCoreError(0x2200000, "test read error"),
				sendSucc: true,
			},
		}
		if err := mgr.Handshake(handshake, peer); err == nil || err.(*core.CoreError).Code() != 0x2200000 {
			t.Errorf("Failed to detect handshake msg read error: %s", err)
			return
		}
	}
}

func TestHandshakeWrongMsg(t *testing.T) {
	log.SetLogLevel(log.NONE)
	defer log.SetLogLevel(log.NONE)
	db, _ := db.NewDatabaseInMem()
	// create a network config with validator function that accept peer connection
	conf := testNetworkConfig(nil, nil, func(config *AppConfig) error {
			return nil
	})
	if mgr, err := NewPlatformManager(&conf.AppConfig, &conf.ServiceConfig, db); err != nil {
		t.Errorf("Failed to create platform manager: %s", err)
	} else {
		// create a peer handshake message with network ID mismatch
		handshake := testPeerHandshakeMsg(mgr.config)
		peer := &mockNode{
			peerNode: peerNode{
				peer: p2p.NewPeer([64]byte(*core.BytesToByte64([]byte("a random peer node ID"))), "", nil),
			},
			testMocks: testMocks{
				readResp: p2p.Msg{
					Payload: &testReader{
						response: []byte("some test response"),
					},
					Code: 0x22222,
				},
				sendSucc: true,
			},
		}
		if err := mgr.Handshake(handshake, peer); err == nil || err.(*core.CoreError).Code() != ErrorHandshakeFailed {
			t.Errorf("Failed to detect incorrect message code: %s", err)
			return
		}
	}
}

func TestHandshakeDecodeError(t *testing.T) {
	log.SetLogLevel(log.NONE)
	defer log.SetLogLevel(log.NONE)
	db, _ := db.NewDatabaseInMem()
	// create a network config with validator function that accept peer connection
	conf := testNetworkConfig(nil, nil, func(config *AppConfig) error {
			return nil
	})
	if mgr, err := NewPlatformManager(&conf.AppConfig, &conf.ServiceConfig, db); err != nil {
		t.Errorf("Failed to create platform manager: %s", err)
	} else {
		// create a peer handshake message with network ID mismatch
		handshake := testPeerHandshakeMsg(mgr.config)
		peer := &mockNode{
			peerNode: peerNode{
				peer: p2p.NewPeer([64]byte(*core.BytesToByte64([]byte("a random peer node ID"))), "", nil),
			},
			testMocks: testMocks{
				readResp: p2p.Msg{
					Payload: &testReader{
						response: nil,
						err: core.NewCoreError(0x232323, "test decode error"),
					},
					Code: Handshake,
				},
				sendSucc: true,
			},
		}
		if err := mgr.Handshake(handshake, peer); err == nil || err.(*core.CoreError).Code() != ErrorHandshakeFailed {
			t.Errorf("Failed to detect message decode error: %s", err)
			return
		}
	}
}

func TestPlatformManagerSubmitTx(t *testing.T) {
	log.SetLogLevel(log.NONE)
	defer log.SetLogLevel(log.NONE)
	db, _ := db.NewDatabaseInMem()
	called := false
	var payload []byte
	var block consensus.Block
	conf := testNetworkConfig(func(tx *Transaction) bool{
			called = true
			payload = append(payload, tx.Payload()...)
			block = tx.block
			return true
		}, nil, nil)
	if mgr, err := NewPlatformManager(&conf.AppConfig, &conf.ServiceConfig, db); err != nil {
		t.Errorf("Failed to create platform manager: %s", err)
	} else {
		// submit transaction
		txPayload := []byte("test tx payload")
		txSubmitter := core.BytesToByte64([]byte("test rx submitter"))
		mgr.Submit(txPayload, txSubmitter)
		// hack, call to processTx should actually be from block producer go routine
		mgr.processTx(<- mgr.txQ, mgr.engine.NewCandidateBlock())
		if !called {
			t.Errorf("transaction never got processed")
		}
		if string(payload) != "test tx payload" {
			t.Errorf("transaction payload incorrect: '%s'", payload)
		}
		if block == nil {
			t.Errorf("transaction block is nil")
		}
	}
}

func TestPlatformManagerBlockProducer(t *testing.T) {
	log.SetLogLevel(log.NONE)
	defer log.SetLogLevel(log.NONE)
	db, _ := db.NewDatabaseInMem()
	called := false
	var payload []byte
	var block consensus.Block
	conf := testNetworkConfig(func(tx *Transaction) bool{
			called = true
			payload = append(payload, tx.Payload()...)
			block = tx.block
			return true
		}, nil, nil)
	if mgr, err := NewPlatformManager(&conf.AppConfig, &conf.ServiceConfig, db); err != nil {
		t.Errorf("Failed to create platform manager: %s", err)
	} else {
		// override timeout
		maxTxWaitSec = 100 * time.Millisecond
		// start block producer
		go mgr.blockProducer()
		// submit transaction
		txPayload := []byte("test tx payload")
		txSubmitter := core.BytesToByte64([]byte("test rx submitter"))
		mgr.Submit(txPayload, txSubmitter)
		// sleep a bit, hoping transaction will get processed till then
		time.Sleep(100 * time.Millisecond)
		mgr.shutdownBlockProducer <- true
		if !called {
			t.Errorf("transaction never got processed")
		}
		if string(payload) != "test tx payload" {
			t.Errorf("transaction payload incorrect: '%s'", payload)
		}
		if block == nil {
			t.Errorf("transaction block is nil")
		}
	}
}

func TestPlatformManagerBlockProducerWaitTimeout(t *testing.T) {
	log.SetLogLevel(log.NONE)
	defer log.SetLogLevel(log.NONE)
	db, _ := db.NewDatabaseInMem()
	called := false
	conf := testNetworkConfig(func(tx *Transaction) bool{
			called = true
			return true
		}, nil, nil)
	if mgr, err := NewPlatformManager(&conf.AppConfig, &conf.ServiceConfig, db); err != nil {
		t.Errorf("Failed to create platform manager: %s", err)
	} else {
		// get current tip from consensus engine
		pre := mgr.engine.BestBlock()
		// override timeout
		maxTxWaitSec = 100 * time.Millisecond
		// start block producer
		go mgr.blockProducer()
		// DO NOT submit transaction
		// sleep a bit longer than timeout, a block should be produced by then
		time.Sleep(150 * time.Millisecond)
		mgr.shutdownBlockProducer <- true
		// get tip after timeout
		post := mgr.engine.BestBlock()
		if post.Depth().Uint64() <= pre.Depth().Uint64() {
			t.Errorf("no blocks produced on timeout")
		}
	}
}

func TestPlatformManagerBlockProducerMultipleTransactions(t *testing.T) {
	log.SetLogLevel(log.NONE)
	defer log.SetLogLevel(log.NONE)
	db, _ := db.NewDatabaseInMem()
	callCount := 0
	conf := testNetworkConfig(func(tx *Transaction) bool{
			callCount++
			return true
		}, nil, nil)
	if mgr, err := NewPlatformManager(&conf.AppConfig, &conf.ServiceConfig, db); err != nil {
		t.Errorf("Failed to create platform manager: %s", err)
	} else {
		// get current tip from consensus engine
		pre := mgr.engine.BestBlock()
		// override timeout
		maxTxWaitSec = 100 * time.Millisecond
		// start block producer
		go mgr.blockProducer()
		// submit multiple transactions transaction
		for i := 0; i<5; i++ {
			txPayload := []byte("test tx payload")
			txSubmitter := core.BytesToByte64([]byte("test rx submitter"))
			mgr.Submit(txPayload, txSubmitter)
		}
		// sleep a bit, one block should be produced by then
		time.Sleep(100 * time.Millisecond)
		mgr.shutdownBlockProducer <- true
		// get tip after timeout
		post := mgr.engine.BestBlock()
		if callCount != 5 {
			t.Errorf("number of processed transactions no correct: %d", callCount)
		}
		if post.Depth().Uint64() != pre.Depth().Uint64()+1 {
			t.Errorf("more than 1 blocks produced")
		}
	}
}

func TestPlatformManagerBlockProducerMaxTransactionsPerBlock(t *testing.T) {
	log.SetLogLevel(log.NONE)
	defer log.SetLogLevel(log.NONE)
	db, _ := db.NewDatabaseInMem()
	callCount := 0
	conf := testNetworkConfig(func(tx *Transaction) bool{
			callCount++
			return true
		}, nil, nil)
	if mgr, err := NewPlatformManager(&conf.AppConfig, &conf.ServiceConfig, db); err != nil {
		t.Errorf("Failed to create platform manager: %s", err)
	} else {
		// get current tip from consensus engine
		pre := mgr.engine.BestBlock()
		// override timeout
		maxTxWaitSec = 100 * time.Millisecond
		// start block producer
		go mgr.blockProducer()
		// submit multiple transactions transaction
		for i := 0; i<15; i++ {
			txPayload := []byte("test tx payload")
			txSubmitter := core.BytesToByte64([]byte("test rx submitter"))
			mgr.Submit(txPayload, txSubmitter)
		}
		// sleep a bit, two block should be produced by then
		time.Sleep(100 * time.Millisecond)
		mgr.shutdownBlockProducer <- true
		// get tip after timeout
		post := mgr.engine.BestBlock()
		if callCount != 15 {
			t.Errorf("number of processed transactions no correct: %d", callCount)
		}
		if post.Depth().Uint64() != pre.Depth().Uint64()+2 {
			t.Errorf("did not produce 2 blocks")
		}
	}
}

func TestPlatformManagerTransactionStatus(t *testing.T) {
	log.SetLogLevel(log.NONE)
	defer log.SetLogLevel(log.NONE)
	db, _ := db.NewDatabaseInMem()
	var block consensus.Block
	conf := testNetworkConfig(func(tx *Transaction) bool{
			block = tx.block
			return true
		}, nil, nil)
	if mgr, err := NewPlatformManager(&conf.AppConfig, &conf.ServiceConfig, db); err != nil {
		t.Errorf("Failed to create platform manager: %s", err)
	} else {
		// override timeout
		maxTxWaitSec = 100 * time.Millisecond
		// start block producer
		go mgr.blockProducer()
		// submit transaction
		txPayload := []byte("test tx payload")
		txSubmitter := core.BytesToByte64([]byte("test rx submitter"))
		txId := mgr.Submit(txPayload, txSubmitter)
		// sleep a bit, hoping transaction will get processed till then
		time.Sleep(100 * time.Millisecond)
		mgr.shutdownBlockProducer <- true
		// query status of the transaction
		if txBlock, err := mgr.Status(txId); err != nil {
			t.Errorf("Failed to get transaction status: %s", err)
		} else if *txBlock.Hash() != *block.Hash() {
			t.Errorf("submitted transaction's block incorrect: %s", err)
		}
	}
}

func TestPlatformManagerUnknownTransactionStatus(t *testing.T) {
	log.SetLogLevel(log.NONE)
	defer log.SetLogLevel(log.NONE)
	db, _ := db.NewDatabaseInMem()
	conf := testNetworkConfig(func(tx *Transaction) bool{
			return true
		}, nil, nil)
	if mgr, err := NewPlatformManager(&conf.AppConfig, &conf.ServiceConfig, db); err != nil {
		t.Errorf("Failed to create platform manager: %s", err)
	} else {
		// override timeout
		maxTxWaitSec = 100 * time.Millisecond
		// start block producer
		go mgr.blockProducer()
		// submit transaction
		txPayload := []byte("test tx payload")
		txSubmitter := core.BytesToByte64([]byte("test rx submitter"))
		mgr.Submit(txPayload, txSubmitter)
		txId := core.BytesToByte64([]byte("some random transaction Id"))
		// sleep a bit, hoping transaction will get processed till then
		time.Sleep(100 * time.Millisecond)
		mgr.shutdownBlockProducer <- true
		// query status of the transaction
		if _, err := mgr.Status(txId); err == nil || err.(*core.CoreError).Code() != consensus.ERR_TX_NOT_FOUND {
			t.Errorf("Failed to detect unknown transaction")
		}
	}
}

func TestPlatformManagerRejectedTransactionStatus(t *testing.T) {
	log.SetLogLevel(log.NONE)
	defer log.SetLogLevel(log.NONE)
	db, _ := db.NewDatabaseInMem()
	conf := testNetworkConfig(func(tx *Transaction) bool{
			return false
		}, nil, nil)
	if mgr, err := NewPlatformManager(&conf.AppConfig, &conf.ServiceConfig, db); err != nil {
		t.Errorf("Failed to create platform manager: %s", err)
	} else {
		// override timeout
		maxTxWaitSec = 100 * time.Millisecond
		// start block producer
		go mgr.blockProducer()
		// submit transaction
		txPayload := []byte("test tx payload")
		txSubmitter := core.BytesToByte64([]byte("test rx submitter"))
		txId := mgr.Submit(txPayload, txSubmitter)
		// sleep a bit, hoping transaction will get processed till then
		time.Sleep(100 * time.Millisecond)
		mgr.shutdownBlockProducer <- true
		// query status of the transaction
		if txBlock, err := mgr.Status(txId); err == nil || err.(*core.CoreError).Code() != consensus.ERR_TX_NOT_APPLIED {
			t.Errorf("Failed to detect rejected transaction")
		} else if txBlock != nil {
			t.Errorf("rejected transaction should have nil block")
		}
	}
}

func TestPlatformManagerPowCallback(t *testing.T) {
	log.SetLogLevel(log.NONE)
	defer log.SetLogLevel(log.NONE)
	db, _ := db.NewDatabaseInMem()
	called := false
	var block consensus.Block
	var hash core.Byte64
	conf := testNetworkConfig(func(tx *Transaction) bool{
			block = tx.block
			return true
		}, func(powHash []byte) bool {
			called = true
			hash = *core.BytesToByte64(powHash)
			return true
		}, nil)
	if mgr, err := NewPlatformManager(&conf.AppConfig, &conf.ServiceConfig, db); err != nil {
		t.Errorf("Failed to create platform manager: %s", err)
	} else {
		// override timeout
		maxTxWaitSec = 100 * time.Millisecond
		// start block producer
		go mgr.blockProducer()
		// submit transaction
		txPayload := []byte("test tx payload")
		txSubmitter := core.BytesToByte64([]byte("test rx submitter"))
		mgr.Submit(txPayload, txSubmitter)
		// sleep a bit, hoping transaction will get processed till then
		time.Sleep(100 * time.Millisecond)
		mgr.shutdownBlockProducer <- true
		// query status of the transaction
		if !called {
			t.Errorf("pow approver never got called")
		}
		if hash != *block.Hash() {
			t.Errorf("pow did not get block's hash")
		}
	}
}

func TestPlatformManagerPowTimeout(t *testing.T) {
	log.SetLogLevel(log.NONE)
	defer log.SetLogLevel(log.NONE)
	db, _ := db.NewDatabaseInMem()
	called := false
	var block consensus.Block
	var hash *core.Byte64
	conf := testNetworkConfig(func(tx *Transaction) bool{
			block = tx.block
			return true
		}, func(powHash []byte) bool {
			called = true
			hash = core.BytesToByte64(powHash)
			return false
		}, nil)
	if mgr, err := NewPlatformManager(&conf.AppConfig, &conf.ServiceConfig, db); err != nil {
		t.Errorf("Failed to create platform manager: %s", err)
	} else {
		log.SetLogLevel(log.DEBUG)
		// override timeout
//		maxTxWaitSec = 100 * time.Millisecond
		// start block producer
		go mgr.blockProducer()
		// submit transaction
		txPayload := []byte("test tx payload")
		txSubmitter := core.BytesToByte64([]byte("test rx submitter"))
		txId := mgr.Submit(txPayload, txSubmitter)
		// sleep a bit, hoping transaction will get processed till then
		time.Sleep(100 * time.Millisecond)
		mgr.shutdownBlockProducer <- true
		// query status of the transaction
		if txBlock, err := mgr.Status(txId); err == nil || err.(*core.CoreError).Code() != consensus.ERR_TX_NOT_APPLIED {
			t.Errorf("Failed to detect rejected transaction")
		} else if txBlock != nil {
			t.Errorf("rejected transaction should have nil block")
		}
		// query status of the transaction
		if !called {
			t.Errorf("pow approver never got called")
		}
		if hash == nil {
			t.Errorf("pow did not get block's hash")
		}
	}
}
