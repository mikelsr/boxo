package bstest

import (
	delay "github.com/ipfs/go-ipfs-delay"
	testinstance "github.com/mikelsr/boxo/bitswap/testinstance"
	tn "github.com/mikelsr/boxo/bitswap/testnet"
	"github.com/mikelsr/boxo/blockservice"
	mockrouting "github.com/mikelsr/boxo/routing/mock"
)

// Mocks returns |n| connected mock Blockservices
func Mocks(n int) []blockservice.BlockService {
	net := tn.VirtualNetwork(mockrouting.NewServer(), delay.Fixed(0))
	sg := testinstance.NewTestInstanceGenerator(net, nil, nil)

	instances := sg.Instances(n)

	var servs []blockservice.BlockService
	for _, i := range instances {
		servs = append(servs, blockservice.New(i.Blockstore(), i.Exchange))
	}
	return servs
}
