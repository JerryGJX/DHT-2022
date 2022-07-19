package chord

import "math/big"

type WrapNode struct {
	node *Node
}

func (ptr *WrapNode) FindSuccessor(target *big.Int, result *string) error {
	createLog(ptr.node.address, "network.RemoteCall", "RemoteCall.FindSuccessor", "Info", "before findSuccessor")
	return ptr.node.findSuccessor(target, result)
}

func (ptr *WrapNode) GetSuccessorList(_ string, result *[successorListSize]string) error {
	return ptr.node.getSuccessorList(result)
}

func (ptr *WrapNode) GetPredecessor(_ string, result *string) error {
	return ptr.node.getPredecessor(result)
}

func (ptr *WrapNode) Notify(addr string, _ *string) error {
	return ptr.node.notify(addr)
}

func (ptr *WrapNode) FixPredecessor(_ string, _ *string) error {
	return ptr.node.fixPredecessor()
}

func (ptr *WrapNode) Stabilize(_ string, _ *string) error {
	ptr.node.stabilize()
	return nil
}

func (ptr *WrapNode) StoreData(dataPair Pair, _ *string) error {
	return ptr.node.storeData(dataPair)
}

func (ptr *WrapNode) GetData(key string, value *string) error {
	return ptr.node.getData(key, value)
}

func (ptr *WrapNode) DeleteData(key string, _ *string) error {
	return ptr.node.deleteData(key)
}

func (ptr *WrapNode) GetDataPartWhenJoin(predeAddr string, dataSet *map[string]string) error {
	return ptr.node.getDataPartWhenJoin(predeAddr, dataSet)
}

func (ptr *WrapNode) BackupStore(dataPair Pair, _ *string) error {
	return ptr.node.backupStore(dataPair)
}

func (ptr *WrapNode) BackupDelete(key string, _ *string) error {
	return ptr.node.backupDelete(key)
}

func (ptr *WrapNode) BackupAdd(dataSet *map[string]string, _ *string) error {
	return ptr.node.backupAdd(dataSet)
}

func (ptr *WrapNode) BackupRemove(dataSet *map[string]string, _ *string) error {
	return ptr.node.backupRemove(dataSet)
}

func (ptr *WrapNode) GenerateBackup(_ string, dataSet *map[string]string) error {
	return ptr.node.generateBackup(dataSet)
}

//func (ptr *WrapNode) Hello(_ string, _ *string) error {
//	println("Hello")
//	createLog(ptr.node.address, "Hello", "default", "Info", "print hello")
//	return nil
//}
