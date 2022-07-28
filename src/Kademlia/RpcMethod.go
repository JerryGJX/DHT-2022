package Kademlia

import (
	"math/big"
)

type RpcNode struct {
	node *Node
}

type FindNodeArg struct {
	ReqAddr AddrType
	Target  *big.Int
}

type FindNodeRep struct {
	ReqAddr AddrType
	RepAddr AddrType
	Content []ClosestListNode
}

type StoreArg struct {
	Key     string
	Value   string
	priType int
	ReqAddr AddrType
}

type FindValueArg struct {
	Key     string
	ReqAddr AddrType
}

type FindValueRep struct {
	ReqAddr AddrType
	RepAddr AddrType
	Content []ClosestListNode
	IfFind  bool
	Value   string
}

func (ptr *RpcNode) FindNode(arg FindNodeArg, result *FindNodeRep) error {
	// if !IfOnline(arg.Requester.Ip){
	// 	return errors.New("requester offline")
	// }
	result.Content = ptr.node.table.FindClosest(arg.Target, K)
	result.ReqAddr = arg.ReqAddr
	result.RepAddr = ptr.node.addr
	ptr.node.table.Update(&arg.ReqAddr)
	return nil
}

func (ptr *RpcNode) Ping(requester AddrType, result *string) error {
	ptr.node.Ping(requester.Ip)
	ptr.node.table.Update(&requester)
	return nil
}

func (ptr *RpcNode) Store(arg StoreArg, result *string) error {
	ptr.node.data.store(arg)
	ptr.node.table.Update(&arg.ReqAddr)
	return nil
}

func (ptr *RpcNode) FindValue(input FindValueArg, result *FindValueRep) error {
	result.Content = ptr.node.table.FindClosest(CalHash(input.Key), K)
	result.ReqAddr = input.ReqAddr
	result.RepAddr = ptr.node.addr
	result.IfFind, result.Value = ptr.node.data.get(input.Key)
	ptr.node.table.Update(&input.ReqAddr)
	return nil
}
