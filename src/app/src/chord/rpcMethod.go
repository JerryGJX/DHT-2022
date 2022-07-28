package chord

import "math/big"

type RpcNode struct {
	node *Node
}

func (ptr *RpcNode) FindSuccessor(target *big.Int, result *string) error {
	createLog(ptr.node.address, "network.RemoteCall", "RemoteCall.FindSuccessor", "Info", "before findSuccessor")
	return ptr.node.findSuccessor(target, result)
}

func (ptr *RpcNode) GetSuccessorList(_ string, result *[successorListSize]string) error {
	return ptr.node.getSuccessorList(result)
}

func (ptr *RpcNode) GetPredecessor(_ string, result *string) error {
	return ptr.node.getPredecessor(result)
}

func (ptr *RpcNode) Notify(addr string, _ *string) error {
	return ptr.node.notify(addr)
}

func (ptr *RpcNode) FixPredecessor(_ string, _ *string) error {
	return ptr.node.fixPredecessor()
}

func (ptr *RpcNode) Stabilize(_ string, _ *string) error {
	ptr.node.stabilize()
	return nil
}

func (ptr *RpcNode) StoreData(dataPair Pair, _ *string) error {
	return ptr.node.dataManagerPtr.storeData(dataPair)
}

func (ptr *RpcNode) GetData(key string, value *string) error {
	return ptr.node.dataManagerPtr.getData(key, value)
}

func (ptr *RpcNode) DeleteData(key string, _ *string) error {
	return ptr.node.dataManagerPtr.deleteData(key)
}

func (ptr *RpcNode) InheritData(preAddr string, dataSet *map[string]string) error {
	return ptr.node.dataManagerPtr.inheritData(preAddr, dataSet)
}

func (ptr *RpcNode) BackupStore(dataPair Pair, _ *string) error {
	return ptr.node.dataManagerPtr.backupStore(dataPair)
}

func (ptr *RpcNode) BackupDelete(key string, _ *string) error {
	return ptr.node.dataManagerPtr.backupDelete(key)
}

func (ptr *RpcNode) BackupAdd(dataSet *map[string]string, _ *string) error {
	return ptr.node.dataManagerPtr.backupAdd(dataSet)
}

func (ptr *RpcNode) BackupRemove(dataSet *map[string]string, _ *string) error {
	return ptr.node.dataManagerPtr.backupRemove(dataSet)
}

func (ptr *RpcNode) GenerateBackup(_ string, dataSet *map[string]string) error {
	return ptr.node.dataManagerPtr.generateBackup(dataSet)
}
