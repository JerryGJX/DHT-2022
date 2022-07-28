package chord

import "math/big"

type WrapNode struct {
	callAddress string
}

func GenerateWrapNode(addr string) WrapNode {
	var ptr WrapNode
	ptr.callAddress = addr
	return ptr
}

func (ptr *WrapNode) FindSuccessor(target *big.Int, result *string) error {
	return RemoteCall(ptr.callAddress, "RpcNode.FindSuccessor", target, result)
}

func (ptr *WrapNode) GetSuccessorList(result *[successorListSize]string) error {
	return RemoteCall(ptr.callAddress, "RpcNode.GetSuccessorList", "", &result)
}

func (ptr *WrapNode) GetPredecessor(result *string) error {
	return RemoteCall(ptr.callAddress, "RpcNode.GetPredecessor", "", &result)
}

func (ptr *WrapNode) Notify(target string) error {
	return RemoteCall(ptr.callAddress, "RpcNode.Notify", target, nil)
}

func (ptr *WrapNode) FixPredecessor() error {
	return RemoteCall(ptr.callAddress, "RpcNode.FixPredecessor", "", nil)
}

func (ptr *WrapNode) Stabilize() error {
	return RemoteCall(ptr.callAddress, "RpcNode.FixPredecessor", "", nil)
}

func (ptr *WrapNode) StoreData(key string, value string) error {
	var keyValuePair Pair
	keyValuePair.Key = key
	keyValuePair.Value = value
	return RemoteCall(ptr.callAddress, "RpcNode.StoreData", keyValuePair, nil)
}

func (ptr *WrapNode) GetData(key string, value *string) error {
	return RemoteCall(ptr.callAddress, "RpcNode.GetData", key, &value)
}

func (ptr *WrapNode) DeleteData(key string) error {
	return RemoteCall(ptr.callAddress, "RpcNode.DeleteData", key, nil)
}

func (ptr *WrapNode) InheritData(preAddr string, preStorage *map[string]string) error {
	return RemoteCall(ptr.callAddress, "RpcNode.InheritData", preAddr, preStorage)
}

func (ptr *WrapNode) BackupStore(data Pair) error {
	return RemoteCall(ptr.callAddress, "RpcNode.BackupStore", data, nil)
}

func (ptr *WrapNode) BackupDelete(key string) error {
	return RemoteCall(ptr.callAddress, "RpcNode.BackupDelete", key, nil)

}

func (ptr *WrapNode) BackupAdd(dataSet *map[string]string) error {
	return RemoteCall(ptr.callAddress, "RpcNode.BackupAdd", dataSet, nil)

}

func (ptr *WrapNode) BackupRemove(dataSet *map[string]string) error {
	return RemoteCall(ptr.callAddress, "RpcNode.BackupRemove", dataSet, nil)
}

func (ptr *WrapNode) GenerateBackup(dataSet *map[string]string) error {
	return RemoteCall(ptr.callAddress, "RpcNode.GenerateBackup", "", dataSet)
}
