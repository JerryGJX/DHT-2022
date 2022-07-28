package chord

import (
	"errors"
	"sync"
)

type DataManager struct {
	nodePtr     *Node
	storage     map[string]string
	storageLock sync.RWMutex
	preBackUp   map[string]string
	backupLock  sync.RWMutex
}

func (ptr *DataManager) Init(targetNode *Node) {
	ptr.nodePtr = targetNode
	ptr.clear()
}

func (ptr *DataManager) inheritData(preAddr string, preStorage *map[string]string) error {
	ptr.backupLock.Lock()
	ptr.storageLock.Lock()
	ptr.preBackUp = make(map[string]string)
	for key, value := range ptr.storage {
		if !In(CalHash(key), CalHash(preAddr), ptr.nodePtr.Id) && key != ptr.nodePtr.address {
			(*preStorage)[key] = value
			ptr.preBackUp[key] = value
			delete(ptr.storage, key)
		}
	}
	ptr.backupLock.Unlock()
	ptr.storageLock.Unlock()
	var sucAddr string
	err := ptr.nodePtr.findFirstOnlineSuccessor(&sucAddr)
	if err != nil {
		return err
	}

	node := GenerateWrapNode(sucAddr)
	_ = node.BackupRemove(preStorage)
	ptr.nodePtr.nodeLock.Lock()
	ptr.nodePtr.predecessor = preAddr
	ptr.nodePtr.nodeLock.Unlock()
	return nil
}

func (ptr *DataManager) backupStore(data Pair) error {
	ptr.backupLock.Lock()
	ptr.preBackUp[data.Key] = data.Value
	ptr.backupLock.Unlock()
	return nil
}

func (ptr *DataManager) backupDelete(key string) error {
	ptr.backupLock.Lock()
	_, flag := ptr.preBackUp[key]
	if flag {
		delete(ptr.preBackUp, key)
	}
	ptr.backupLock.Unlock()
	return nil
}

func (ptr *DataManager) backupAdd(targetPair *map[string]string) error {
	ptr.backupLock.Lock()
	for key, value := range *targetPair {
		ptr.preBackUp[key] = value
	}
	ptr.backupLock.Unlock()
	return nil
}

func (ptr *DataManager) backupRemove(targetPair *map[string]string) error {
	ptr.backupLock.Lock()
	for key := range *targetPair {
		delete(ptr.preBackUp, key)
	}
	ptr.backupLock.Unlock()
	return nil
}

func (ptr *DataManager) generateBackup(sucBackup *map[string]string) error {
	ptr.storageLock.RLock()
	*sucBackup = make(map[string]string)
	for key, value := range ptr.storage {
		(*sucBackup)[key] = value
	}
	ptr.storageLock.RUnlock()
	return nil
}

func (ptr *DataManager) applyBackup() error {
	ptr.backupLock.RLock()
	ptr.storageLock.Lock()
	for key, value := range ptr.preBackUp {
		ptr.storage[key] = value
	}
	ptr.backupLock.RUnlock()
	ptr.storageLock.Unlock()
	var sucAddr string
	err := ptr.nodePtr.findFirstOnlineSuccessor(&sucAddr)
	if err != nil {
		return err
	}

	node := GenerateWrapNode(sucAddr)
	err = node.BackupAdd(&ptr.preBackUp)

	if err != nil {
		return err
	}
	ptr.backupLock.Lock()
	ptr.preBackUp = make(map[string]string)
	ptr.backupLock.Unlock()
	return nil
}

func (ptr *DataManager) storeData(data Pair) error {
	createLog(ptr.nodePtr.address, "Node.storeData", "default", "Info", "into store Data")

	ptr.storageLock.Lock()
	ptr.storage[data.Key] = data.Value
	ptr.storageLock.Unlock()

	var sucAddr string
	err := ptr.nodePtr.findFirstOnlineSuccessor(&sucAddr)

	node := GenerateWrapNode(sucAddr)
	err = node.BackupStore(data)
	if err != nil {
	}
	return err
}

func (ptr *DataManager) getData(key string, value *string) error {
	createLog(ptr.nodePtr.address, "Node.getData", "default", "Info", "enter getData")
	ptr.storageLock.RLock()
	tmp, flag := ptr.storage[key]
	ptr.storageLock.RUnlock()
	if flag {
		*value = tmp
		createLog(ptr.nodePtr.address, "Node.getData", "default", "Info", "get Data")
		return nil
	} else {
		*value = ""
		return errors.New("<getData> not found")
	}
}

func (ptr *DataManager) deleteData(key string) error {
	ptr.storageLock.Lock()
	_, flag := ptr.storage[key]
	if flag {
		delete(ptr.storage, key)
	}
	ptr.storageLock.Unlock()
	if flag {
		var sucAddr string
		_ = ptr.nodePtr.findFirstOnlineSuccessor(&sucAddr)

		node := GenerateWrapNode(sucAddr)
		_ = node.BackupDelete(key)
		return nil
	} else {
		return errors.New("<deleteData> not found")
	}
}

func (ptr *DataManager) clear() {
	ptr.storageLock.Lock()
	ptr.storage = make(map[string]string)
	ptr.storageLock.Unlock()
	ptr.backupLock.Lock()
	ptr.preBackUp = make(map[string]string)
	ptr.backupLock.Unlock()
}
