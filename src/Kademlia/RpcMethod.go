package Kademlia

import (
	"math/big"
)

type RpcNode struct {
	node *Node
}

type FindNodeArg struct {
	Requester AddrType
	Target    *big.Int
}

type FindNodeRep struct {
	Requester AddrType
	Replier   AddrType
	Content   []ContactRecord
}

func (ptr *RpcNode) FindNode(arg FindNodeArg, result *FindNodeRep) error {
	result.Content = ptr.node.table.FindClosest(arg.Target, K)
	result.Requester = arg.Requester
	result.Replier = ptr.node.addr
	ptr.node.table.Update(&arg.Requester)
	return nil
}

func (ptr *RpcNode) Ping(requester AddrType, result *string) error {
	ptr.node.Ping(requester.Ip)
	ptr.node.table.Update(&requester)
	return nil
}

type StoreArg struct {
	Key          string
	Value        string
	RequesterPri int
	Requester    AddrType
}

func (ptr *RpcNode) Store(arg StoreArg, result *string) error {
	ptr.node.data.store(arg)
	ptr.node.table.Update(&arg.Requester)
	return nil
}

type FindValueArg struct {
	Key       string
	Requester AddrType
}

type FindValueRep struct {
	Requester AddrType
	Replier   AddrType
	Content   []ContactRecord
	IsFind    bool
	Value     string
}

func (ptr *RpcNode) FindValue(input FindValueArg, result *FindValueRep) error {
	result.Content = ptr.node.table.FindClosest(CalHash(input.Key), K)
	result.Requester = input.Requester
	result.Replier = ptr.node.addr
	result.IsFind, result.Value = ptr.node.data.get(input.Key)
	ptr.node.table.Update(&input.Requester)
	return nil
}
