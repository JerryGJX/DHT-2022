package Kademlia


type WrapNode struct {
	nodePtr     *Node
	callAddress string
}

func GenerateWrapNode(self *Node, addr string) WrapNode {
	var ptr WrapNode
	ptr.nodePtr = self
	ptr.callAddress = addr
	return ptr
}


func (ptr *WrapNode) Store(arg *StoreArg) error {
	var occupy string
	return RemoteCall(ptr.nodePtr, ptr.callAddress, "RpcNode.Store", arg, &occupy)
}


func (ptr *WrapNode) FindNode(arg *FindNodeArg, rep *FindNodeRep) error {
	return RemoteCall(ptr.nodePtr, ptr.callAddress, "RpcNode.FindNode", arg, rep)
}


func (ptr *WrapNode) FindValue(arg *FindValueArg, rep *FindValueRep) error {
	return RemoteCall(ptr.nodePtr, ptr.callAddress, "RpcNode.FindValue", arg, rep)
}

func (ptr *WrapNode) Ping(targetAddr AddrType) error {
	var occupy string
	return RemoteCall(ptr.nodePtr, ptr.callAddress, "RpcNode.Ping", targetAddr, &occupy)
}
