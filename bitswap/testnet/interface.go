package bitswap

import (
	bsnet "github.com/mikelsr/boxo/bitswap/network"

	tnet "github.com/mikelsr/go-libp2p-testing/net"
	"github.com/mikelsr/go-libp2p/core/peer"
)

// Network is an interface for generating bitswap network interfaces
// based on a test network.
type Network interface {
	Adapter(tnet.Identity, ...bsnet.NetOpt) bsnet.BitSwapNetwork

	HasPeer(peer.ID) bool
}
