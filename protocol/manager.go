package protocol

import (
	"github.com/trust-net/go-trust-net/db"
	"github.com/trust-net/go-trust-net/common"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
)

type ProtocolManager interface {
	// provide an instance of p2p protocol implementation
	Protocol() p2p.Protocol
	
	// initiate connection and handshake with a node
	AddPeer(node *discover.Node) error

	// perform protocol specific handshake with newly connected peer
	Handshake(peer *p2p.Peer, status *HandshakeMsg, ws p2p.MsgReadWriter) error
	
	// get reference to protocol manager's DB
	Db() db.PeerSetDb
	
	// Shutdown and cleanup
	Shutdown()
}

// protocol errors
const (
	ErrorHandshakeFailed = 0x01
	ErrorMaxPeersReached = 0x02
	ErrorUnknownMessageType = 0x03
	ErrorNotImplemented = 0x04
	ErrorSyncFailed = 0x05
	ErrorInvalidRequest = 0x06
	ErrorInvalidResponse = 0x07
	ErrorNotFound = 0x08
	ErrorBadBlock = 0x09
)

// base protocol manager implementation for shared data and code,
// will be extended by actual protocol manager implementations
type ManagerBase struct {
	db db.PeerSetDb
	peerCount	int
}


func (mgr *ManagerBase) PeerCount() int {
	return mgr.peerCount
}

func (mgr *ManagerBase) DecrPeer() {
	mgr.peerCount--
}

func (mgr *ManagerBase) Db() db.PeerSetDb {
	return mgr.db
}

func (mgr *ManagerBase) SetDb(db db.PeerSetDb) {
	mgr.db = db
}

func (mgr *ManagerBase) UnregisterPeer(node *Node) {
	mgr.db.UnRegisterPeerNodeForId(node.Peer().ID().String())
	mgr.DecrPeer()
}

func (mgr *ManagerBase) AddPeer(node *discover.Node) error {
	// we don't have a p2p server for individual protocol manager, and hence cannot add a node
	// this will need to be done from outside, at the application level
	return NewProtocolError(ErrorNotImplemented, "protocol manager cannot add peer")
}

// perform sub protocol handshake
func (mgr *ManagerBase) Handshake(status *HandshakeMsg, peer *Node) error {
	// send our status to the peer
	if err := peer.Send(Handshake, *status); err != nil {
		return NewProtocolError(ErrorHandshakeFailed, err.Error())
	}

	var msg p2p.Msg
	var err error
	err = common.RunTimeBound(5, func() error {
			msg, err = peer.ReadMsg()
			return err
		}, NewProtocolError(ErrorHandshakeFailed, "timed out waiting for handshake status"))
	if err != nil {
		return err
	}

	// make sure its a handshake status message
	if msg.Code != Handshake {
		return NewProtocolError(ErrorHandshakeFailed, "first message needs to be handshake status")
	}
	var handshake HandshakeMsg
	err = msg.Decode(&handshake)
	if err != nil {
		return NewProtocolError(ErrorHandshakeFailed, err.Error())
	}
	
	// validate handshake message
	switch {
		case handshake.NetworkId != status.NetworkId:
			return NewProtocolError(ErrorHandshakeFailed, "network ID does not match")
		case handshake.ShardId != status.ShardId:
			return NewProtocolError(ErrorHandshakeFailed, "shard ID does not match")
	}

	// add the peer into our DB
	if err = mgr.db.RegisterPeerNode(peer); err != nil {
		return err
	} else {
		mgr.peerCount++
		peer.SetStatus(&handshake)
	}
	return nil
}
