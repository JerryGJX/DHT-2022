package chord

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"math/big"
	"sync"
	"time"
)

const (
	successorListSize int = 5
	hashBitsSize      int = 160
)

type Node struct {
	address       string
	Id            *big.Int
	next          int
	predecessor   string
	readyOnline   bool
	quitSignal    chan bool
	successorList [successorListSize]string
	fingerTable   [hashBitsSize]string
	nodeLock      sync.RWMutex //to control all the adjustment for the node
	storage       map[string]string
	storageLock   sync.RWMutex
	preBackUp     map[string]string
	backupLock    sync.RWMutex
	server        *network
}

var printOnce bool

func (ptr *Node) Init(port int) {
	var addr string
	addr = fmt.Sprintf("%s:%d", localAddr, port)
	ptr.address = addr
	ptr.Id = CalHash(addr)
	ptr.readyOnline = false
	ptr.next = 1
	ptr.clear()
}

// Run /* "Run" is called after calling "NewNode". */
func (ptr *Node) Run() {
	createLog(ptr.address, "Node.Run", "default", "Info", "get in")
	ptr.server = new(network)
	err := ptr.server.Init(ptr.address, ptr)
	if err != nil {
		//log.Errorln("<Run> failed with err: ", err)
		createLog(ptr.address, "Node.Run", "network.Init", "Error", err.Error())
		return
	}
	//log.Infoln("<Run> success in address: ", ptr.address)
	createLog(ptr.address, "Node.Run", "default", "Info", "")
	ptr.readyOnline = true
	ptr.next = 1
}

// Create /* "Create" and "Join" are called after calling "Run". */
/* For a dhtNode, either "Create" or "Join" will be called, but not both. */
func (ptr *Node) Create() {
	createLog(ptr.address, "Node.Create", "default", "Info", "get in")
	ptr.predecessor = ""
	ptr.successorList[0] = ptr.address
	ptr.fingerTable[0] = ptr.address
	ptr.update()
	//log.Infoln("<Create> new ring success in ", ptr.address)
	createLog(ptr.address, "Node.Create", "default", "Info", "")
} /* Create a new network. */

func (ptr *Node) Join(addr string) bool {
	createLog(ptr.address, "Node.Join", "default", "Info", "get in")
	if !IfOnline(addr) {
		//log.Warningln("<Join> Node ", addr, " fail to join for ", ptr.address, " is off line")
		createLog(ptr.address, "Node.Join", "network.IfOnline", "Warning", addr+" is off line")
		return false
	}
	createLog(ptr.address, "Node.Join", "default", "Info", "after ifOnline, addr= "+addr)
	var succAddr string
	err := RemoteCall(addr, "WrapNode.FindSuccessor", ptr.Id, &succAddr)
	createLog(ptr.address, "Node.Join", "default", "Info", "after RemoteCall.FindSuccessor")
	if err != nil {
		//log.Errorln("<Join> fail to join for error1: ", err)
		createLog(ptr.address, "Node.Join", "RemoteCall.FindSuccessor", "Error", err.Error())
		return false
	}
	//log.Infoln("<Join> Get Successor and Join Succeed")
	createLog(ptr.address, "Node.Join", "default", "Info", "")
	var tmp [successorListSize]string
	err = RemoteCall(succAddr, "WrapNode.GetSuccessorList", "", &tmp)
	if err != nil {
		//log.Errorln("<Join> fail to get successorList for ", err)
		createLog(ptr.address, "Node.Join", "RemoteCall.GetSuccessorList", "Error", err.Error())
		return false
	}
	ptr.nodeLock.Lock()
	ptr.predecessor = ""
	ptr.successorList[0] = succAddr
	ptr.fingerTable[0] = succAddr
	for i := 1; i < successorListSize; i++ {
		ptr.successorList[i] = tmp[i-1]
	}
	ptr.nodeLock.Unlock()
	err = RemoteCall(succAddr, "WrapNode.GetDataPartWhenJoin", ptr.address, &ptr.storage)
	if err != nil {
		//log.Errorln("<Join> fail to join, cannot GetDataPartWhenJoin for ", err)
		createLog(ptr.address, "Node.Join", "RemoteCall.GetDataPartWhenJoin", "Error", err.Error())
		return false
	}
	ptr.update()
	return true
} /* Join an existing network. Return "true" if join succeeded and "false" if not. */

// Quit /* Quit from the network it is currently in.*/
/* "Quit" will not be called before "Create" or "Join". */
/* For a dhtNode, "Quit" may be called for many times. */
/* For a quited node, call "Quit" again should have no effect. */
func (ptr *Node) Quit() {
	if !ptr.readyOnline {
		return
	}
	//log.Warningln("<Quit> start :", ptr.address)
	//createLog(ptr.address,"Node.Quit","default")

	err := ptr.server.ShutDown()
	if err != nil {
		//log.Errorln("<Quit> fail when server shutdown for ", err)
		createLog(ptr.address, "Node.Quit", "network.ShutDown", "Error", err.Error())
	}
	ptr.nodeLock.Lock()
	ptr.readyOnline = false
	ptr.nodeLock.Unlock()

	var occupy string
	var succAddr string
	err = ptr.findFirstOnlineSuccessor(&succAddr)
	err = RemoteCall(succAddr, "WrapNode.FixPredecessor", "", &occupy)
	if err != nil {
		//log.Errorln("<Quit.FixPredecessor> Error : ", err)
		createLog(ptr.address, "Node.Quit", "RemoteCall.FixPredecessor", "Error", err.Error())
	}
	err = RemoteCall(succAddr, "WrapNode.Stabilize", "", &occupy)
	if err != nil {
		//log.Errorln("<Quit> fail when stabilize for ", err)
		createLog(ptr.address, "Node.Quit", "RemoteCall.Stabilize", "Error", err.Error())
	}
	ptr.clear()
	log.Infoln("<Quit> succeed")
	createLog(ptr.address, "Node.Quit", "default", "Info", "")
}

// ForceQuit /* Chord offers a way of "normal" quitting. */
/* For "force quit", the node quit the network without informing other nodes. */
/* "ForceQuit" will be checked by TA manually. */
func (ptr *Node) ForceQuit() {
	//log.Warningln("<Quit> start :", ptr.address)
	err := ptr.server.ShutDown()
	if err != nil {
		//log.Errorln("<Quit> fail when server shutdown for ", err)
		createLog(ptr.address, "Node.ForceQuit", "network.ShutDown", "Error", err.Error())
	}
	ptr.nodeLock.Lock()
	ptr.readyOnline = false
	ptr.nodeLock.Unlock()

	ptr.clear()
	//log.Infoln("<Quit> succeed")
	createLog(ptr.address, "Node.ForceQuit", "default", "Info", "")
}

// Ping /* Check whether the node represented by the IP address is in the network. */
func (ptr *Node) Ping(addr string) bool {
	var address string
	address = addr
	return IfOnline(address)
}

// Put /* Put a key-value pair into the network (if KEY is already in the network, cover it), or
func (ptr *Node) Put(key string, value string) bool {
	if !ptr.readyOnline {
		//log.Errorln("<Put> the node ", ptr.address, " isn't ready ")
		createLog(ptr.address, "Node.Put", "default", "Error", "self off line")
		return false
	}
	var targetAddr string
	err := ptr.findSuccessor(CalHash(key), &targetAddr)
	if err != nil {
		//log.Errorln("<Put> fail to find key's successor")
		createLog(ptr.address, "Node.Put", "Node.findSuccessor", "Error", err.Error())
		return false
	}
	var occupy string
	var keyValuePair Pair
	keyValuePair.Key = key
	keyValuePair.Value = value
	err = RemoteCall(targetAddr, "WrapNode.StoreData", keyValuePair, &occupy)
	if err != nil {
		//log.Errorln("<Put> fail to put data into target Node ", targetAddr, " for ", err)
		createLog(ptr.address, "Node.Put", "RemoteCall.StoreData", "Error", err.Error())
		return false
	}
	//log.Infoln("<Put> succeed to put into ", targetAddr)
	createLog(ptr.address, "Node.Put", "default", "Info", "")
	return true
} /* Return "true" if success, "false" otherwise. */

func (ptr *Node) Get(key string) (bool, string) {
	//println("GGGGGet")
	createLog(ptr.address, "Node.Get", "default", "Info", "enter Node.Get")
	if !ptr.readyOnline {
		//log.Errorln("<Get> the node ", ptr.address, " isn't ready ")
		createLog(ptr.address, "Node.Get", "default", "Error", "self off line")
		return false, ""
	}
	var value string
	var targetAddr string
	err := ptr.findSuccessor(CalHash(key), &targetAddr)
	if err != nil {
		//log.Errorln("<Get> fail to find key's successor")
		createLog(ptr.address, "Node.Get", "Node.findSuccessor", "Error", err.Error())
		return false, ""
	}
	err = RemoteCall(targetAddr, "WrapNode.GetData", key, &value)
	if err != nil {
		//log.Errorln("<Get> fail to put data into target Node ", targetAddr, " for ", err)
		createLog(ptr.address, "Node.Get", "Node.GetData", "Error", err.Error())
		return false, ""
	}
	//log.Infoln("<Get> succeed to get form ", targetAddr)
	createLog(ptr.address, "Node.Get", "default", "Info", "")
	return true, value
} /* Return "true" and the value if success, "false" otherwise. */

func (ptr *Node) Delete(key string) bool {
	if !ptr.readyOnline {
		//log.Errorln("<Delete> the node ", ptr.address, " isn't ready ")
		createLog(ptr.address, "Node.Delete", "default", "Error", "self off line")
		return false
	}
	var targetAddr string
	err := ptr.findSuccessor(CalHash(key), &targetAddr)
	if err != nil {
		//log.Errorln("<Delete> fail to find key's successor")
		createLog(ptr.address, "Node.Delete", "Node.findSuccessor", "Error", err.Error())
		return false
	}
	var occupy string
	err = RemoteCall(targetAddr, "WrapNode.DeleteData", key, &occupy)
	if err != nil {
		//log.Errorln("<Delete> fail to delete data in target Node ", targetAddr, " for ", err)
		createLog(ptr.address, "Node.Delete", "Node.DeleteData", "Error", err.Error())
		return false
	}
	//log.Infoln("<> succeed to put into ", targetAddr)
	createLog(ptr.address, "Node.Delete", "default", "Info", "")
	return true
} /* Remove the key-value pair represented by KEY from the network. */
/* Return "true" if remove successfully, "false" otherwise. */

func (ptr *Node) DataSize() int {
	return len(ptr.storage)
}

//--------------------private func----------------------
func createLog(addr string, callFuncName string, funcName string, logType string, info string) {
	if logType == "Info" {
		log.Infof("[%s] <%s> Call <%s> Succeed %s\n", addr, callFuncName, funcName, info)
	} else if logType == "Warning" {
		log.Warningf("[%s] <%s> Call <%s> Warning %s\n", addr, callFuncName, funcName, info)
	} else if logType == "Error" {
		log.Errorf("[%s] <%s> Call <%s> Error %s\n", addr, callFuncName, funcName, info)
	}
}

func (ptr *Node) findSuccessor(targetId *big.Int, result *string) error {

	var succAddr string
	err := ptr.findFirstOnlineSuccessor(&succAddr)
	//if err != nil {
	//	return err
	//}
	succId := CalHash(succAddr)

	//println("succAddr", succAddr)
	//println("targetId", targetId)
	//println("succId", succId)
	//println("ptrId", ptr.Id)

	var tmp Identifier
	tmp.Value = targetId

	if targetId.Cmp(succId) == 0 || tmp.In(ptr.Id, succId) {
		//createLog(ptr.address, "Node.findSuccessor", "default", "Info", "get successor")
		*result = succAddr
		return nil
	}
	nextStep := ptr.closestPrecedingFinger(targetId)
	if err != nil {

	}
	createLog(ptr.address, "node.findSuccessor", "dafault", "Info", "findSucc last")
	return RemoteCall(nextStep, "WrapNode.FindSuccessor", targetId, result)

}

func (ptr *Node) closestPrecedingFinger(targetId *big.Int) string {
	for i := hashBitsSize - 1; i >= 0; i-- {
		if ptr.fingerTable[i] == "" || !IfOnline(ptr.fingerTable[i]) {
			continue
		}
		fingerId := GenerateId(ptr.fingerTable[i])
		if fingerId.In(ptr.Id, targetId) {
			//log.Infoln("<closestPrecedingFinger> Find closestPrecedingFinger Successfully in Node ", ptr.address)
			return ptr.fingerTable[i]
		}
	}
	var preAddr string
	err := ptr.findFirstOnlineSuccessor(&preAddr)
	if err != nil {
		//log.Errorln("<closestPrecedingFinger> list break")
		return ""
	}

	return preAddr
}

func (ptr *Node) findFirstOnlineSuccessor(result *string) error {
	createLog(ptr.address, "findFirstOnlineSuccessor", "default", "Info", "into find first online successor")

	for i := 0; i < successorListSize; i++ {
		if IfOnline(ptr.successorList[i]) {
			//println("####")
			//createLog(ptr.address, "findFirstOnlineSuccessor", "default", "Waring", "find online")
			*result = ptr.successorList[i]
			//println(i, "#", ptr.successorList[i])
			return nil
		}
		createLog(ptr.address, "findFirstOnlineSuccessor", "default", "Waring", "didn't find online")
	}
	//log.Errorln("<find_first_online_successor> List Break in ", ptr.address)

	return errors.New("list Break")
}

func (ptr *Node) clear() {
	ptr.storageLock.Lock()
	ptr.storage = make(map[string]string)
	ptr.storageLock.Unlock()

	ptr.backupLock.Lock()
	ptr.preBackUp = make(map[string]string)
	ptr.backupLock.Unlock()

	ptr.nodeLock.Lock()
	ptr.quitSignal = make(chan bool, 2)
	ptr.next = 1
	ptr.nodeLock.Unlock()
}

func (ptr *Node) getPredecessor(result *string) error {
	ptr.nodeLock.RLock()
	*result = ptr.predecessor
	ptr.nodeLock.RUnlock()
	return nil
}

func (ptr *Node) getSuccessorList(result *[successorListSize]string) error {
	ptr.nodeLock.RLock()
	*result = ptr.successorList
	ptr.nodeLock.RUnlock()
	return nil
}

func (ptr *Node) stabilize() {
	var succPreAddr, newSuccAddr string
	err := ptr.findFirstOnlineSuccessor(&newSuccAddr)
	err = RemoteCall(newSuccAddr, "WrapNode.GetPredecessor", "", &succPreAddr)
	if err != nil {
		//log.Errorln("<stabilize> fail to get predecessor of ", newSuccAddr)
		return
	}
	id := GenerateId(succPreAddr)
	if succPreAddr != "" && id.In(ptr.Id, CalHash(newSuccAddr)) {
		newSuccAddr = succPreAddr
	}
	var tmp [successorListSize]string
	err = RemoteCall(newSuccAddr, "WrapNode.GetSuccessorList", "A", &tmp)
	if err != nil {
		//log.Errorln("<stabilize> fail to get successorList for ", err)
		return
	}
	ptr.nodeLock.Lock()
	ptr.successorList[0] = newSuccAddr
	ptr.fingerTable[0] = newSuccAddr
	for i := 1; i < successorListSize; i++ {
		ptr.successorList[i] = tmp[i-1]
	}
	ptr.nodeLock.Unlock()
	var occupy string
	err = RemoteCall(newSuccAddr, "WrapNode.Notify", ptr.address, &occupy)
	if err != nil {
		//log.Errorln("<stabilize> Fail to call successor for ", err)
		return
	}
	//log.Infoln("<stabilize> successfully :) in ", ptr.address)
}

func (ptr *Node) notify(addr string) error {
	if ptr.predecessor == addr {
		return nil
	}

	id := GenerateId(addr)
	if ptr.predecessor == "" || id.In(CalHash(ptr.predecessor), ptr.Id) {
		ptr.nodeLock.Lock()
		ptr.predecessor = addr
		ptr.nodeLock.Unlock()
		//for backup
		var occupy string
		err := RemoteCall(addr, "WrapNode.GenerateBackup", occupy, &ptr.preBackUp)
		if err != nil {
			//log.Errorln("<notify> Fail to get predecessor ", addr, " generate backup for ", err)
		}
		//log.Errorln("<notify> Succeed in getting predecessor ", addr, " generate backup")
	}
	return nil
}

func (ptr *Node) fixPredecessor() error {
	if ptr.predecessor != "" && !IfOnline(ptr.predecessor) {
		ptr.nodeLock.Lock()
		ptr.predecessor = ""
		ptr.nodeLock.Unlock()
		err := ptr.applyBackup()
		if err != nil {
			return err
		}
		//log.Infoln("<fixPredecessor> find failed predecessor")
	}
	return nil
}

func (ptr *Node) fixFinger() {
	var succAddr string
	err := ptr.findSuccessor(PlusTwoPower(ptr.Id, ptr.next), &succAddr)
	if err != nil {
		createLog(ptr.address, "fixFinger", "findSuccessor", "Error", err.Error())
		return
	}
	ptr.nodeLock.Lock()
	ptr.fingerTable[0] = ptr.successorList[0]
	ptr.fingerTable[ptr.next] = succAddr
	ptr.next += 1
	if ptr.next >= hashBitsSize {
		ptr.next = 1
	}
	ptr.nodeLock.Unlock()
	createLog(ptr.address, "fixFinger", "default", "Info", "")
}

func (ptr *Node) update() {
	go func() {
		for ptr.readyOnline {
			ptr.stabilize()
			time.Sleep(CutTime)
		}
	}()
	go func() {
		for ptr.readyOnline {
			ptr.fixPredecessor()
			time.Sleep(CutTime)
		}
	}()
	go func() {
		for ptr.readyOnline {
			ptr.fixFinger()
			time.Sleep(CutTime)
		}
	}()
}

//when joining, get part of successor data move to it's predecessor
func (ptr *Node) getDataPartWhenJoin(preAddr string, preStorage *map[string]string) error {
	ptr.backupLock.Lock()
	ptr.storageLock.Lock()
	ptr.preBackUp = make(map[string]string)
	var keyValuePair Pair
	for key, value := range ptr.storage {
		keyValuePair.Key = key
		keyValuePair.Value = value

		id := GenerateId(key)
		if !id.InRightClose(CalHash(preAddr), ptr.Id) && keyValuePair.Key != ptr.address {
			(*preStorage)[key] = value
			ptr.preBackUp[key] = value
			delete(ptr.storage, key)
		}
	}
	ptr.backupLock.Unlock()
	ptr.storageLock.Unlock()
	var succAddr string
	err := ptr.findFirstOnlineSuccessor(&succAddr)
	if err != nil {
		return err
	}
	var blank string
	err = RemoteCall(succAddr, "WrapNode.BackupRemove", &preStorage, &blank)
	if err != nil {
		//log.Errorln("<getDataPartWhenJoin> fail to remove pairs in successor for ", err)
	}
	ptr.nodeLock.Lock()
	ptr.predecessor = preAddr
	ptr.nodeLock.Unlock()
	//log.Infoln("<getDataPartWhenJoin> succeed in passing pairs to successor by ", preAddr)
	return nil
}

func (ptr *Node) backupStore(data Pair) error {
	println("In to backupStore")
	ptr.backupLock.Lock()
	ptr.preBackUp[data.Key] = data.Value
	ptr.backupLock.Unlock()
	println("exit backupStore")
	return nil
}

func (ptr *Node) backupDelete(key string) error {
	ptr.backupLock.Lock()
	delete(ptr.preBackUp, key)
	return nil
}

func (ptr *Node) backupAdd(targetPair *map[string]string) error {
	ptr.backupLock.Lock()
	for key, value := range *targetPair {
		ptr.preBackUp[key] = value
	}
	ptr.backupLock.Unlock()
	return nil
}

func (ptr *Node) backupRemove(targetPair *map[string]string) error {
	ptr.backupLock.Lock()
	for key := range *targetPair {
		delete(ptr.preBackUp, key)
	}
	ptr.backupLock.Unlock()
	return nil
}

func (ptr *Node) generateBackup(sucBackup *map[string]string) error {
	ptr.storageLock.RLock()
	*sucBackup = make(map[string]string)
	for key, value := range ptr.storage {
		(*sucBackup)[key] = value
	}
	ptr.storageLock.RUnlock()
	return nil
}

func (ptr *Node) applyBackup() error {
	ptr.backupLock.RLock()
	ptr.storageLock.Lock()
	for key, value := range ptr.preBackUp {
		ptr.storage[key] = value
	}
	ptr.backupLock.RUnlock()
	ptr.storageLock.Unlock()
	var succAddr string
	err := ptr.findFirstOnlineSuccessor(&succAddr)
	if err != nil {
		//log.Errorln("<applyBackup> fail to find online successor of ", ptr.address)
		return err
	}
	var occupy string
	err = RemoteCall(succAddr, "WrapNode.BackupAdd", ptr.preBackUp, &occupy)
	if err != nil {
		//log.Errorln("<applyBackup> fail to update ", ptr.address, " successor's preBackup")
		return err
	}
	ptr.backupLock.Lock()
	ptr.preBackUp = make(map[string]string)
	ptr.backupLock.Unlock()
	//log.Infoln("<applyBackup> success in address ", ptr.address)
	return nil
}

func (ptr *Node) storeData(data Pair) error {
	createLog(ptr.address, "Node.storeData", "default", "Info", "into store Data")

	println("data.key", data.Key)
	println("data.value", data.Value)
	ptr.storageLock.Lock()
	ptr.storage[data.Key] = data.Value
	ptr.storageLock.Unlock()

	println("storage size", len(ptr.storage))

	var succAddr string
	err := ptr.findFirstOnlineSuccessor(&succAddr)
	var occupy string
	err = RemoteCall(succAddr, "WrapNode.BackupStore", data, &occupy)
	if err != nil {
		//log.Errorln("<storeData> fail to store backup in ", ptr.address)
	}
	return err
}

func (ptr *Node) getData(key string, value *string) error {
	//createLog(ptr.address, "Node.getData", "default", "Info", "enter getData")
	ptr.storageLock.RLock()
	tmp, flag := ptr.storage[key]
	ptr.storageLock.RUnlock()
	if flag {
		*value = tmp
		createLog(ptr.address, "Node.getData", "default", "Info", "get Data")
		return nil
	} else {
		*value = ""
		return errors.New("<getData> not found")
	}
}

func (ptr *Node) deleteData(key string) error {
	ptr.storageLock.Lock()
	_, flag := ptr.storage[key]
	if flag {
		delete(ptr.storage, key)
	}
	ptr.storageLock.Unlock()
	if flag {
		var succAddr string
		err := ptr.findFirstOnlineSuccessor(&succAddr)
		var occupy string
		err = RemoteCall(succAddr, "WrapNode.BackupDelete", key, &occupy)
		if err != nil {
			//log.Errorln("<deleteData> fail to delete backup in ", ptr.address)
		}
		return nil
	} else {
		return errors.New("<deleteData> not found")
	}
}
