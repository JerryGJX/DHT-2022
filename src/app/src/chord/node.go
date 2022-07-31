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
	dataManagerPtr *DataManager
	address        string
	Id             *big.Int
	next           int
	predecessor    string
	readyOnline    bool
	quitSignal     chan bool
	successorList  [successorListSize]string
	fingerTable    [hashBitsSize]string
	nodeLock       sync.RWMutex //to control all the adjustment for the node
	server         *network
}

func (ptr *Node) Init(port int) {
	var addr string
	addr = fmt.Sprintf("%s:%d", localAddr, port)
	ptr.address = addr
	ptr.Id = CalHash(addr)
	ptr.readyOnline = false
	ptr.dataManagerPtr = new(DataManager)
	ptr.dataManagerPtr.Init(ptr)
	ptr.clear()
}

// Run /* "Run" is called after calling "NewNode". */
func (ptr *Node) Run() {
	createLog(ptr.address, "Node.Run", "default", "Info", "get in")
	ptr.server = new(network)
	err := ptr.server.Init(ptr.address, ptr)
	if err != nil {
		createLog(ptr.address, "Node.Run", "network.Init", "Error", err.Error())
		return
	}
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
	createLog(ptr.address, "Node.Create", "default", "Info", "")
} /* Create a new network. */

func (ptr *Node) Join(addr string) bool {
	createLog(ptr.address, "Node.Join", "default", "Info", "get in")
	if !IfOnline(addr) {
		createLog(ptr.address, "Node.Join", "network.IfOnline", "Warning", addr+" is off line")
		return false
	}
	createLog(ptr.address, "Node.Join", "default", "Info", "after ifOnline, addr= "+addr)

	var sucAddr string
	node := GenerateWrapNode(addr)
	err := node.FindSuccessor(ptr.Id, &sucAddr)

	createLog(ptr.address, "Node.Join", "default", "Info", "after RemoteCall.FindSuccessor")
	if err != nil {
		createLog(ptr.address, "Node.Join", "RemoteCall.FindSuccessor", "Error", err.Error())
		return false
	}
	createLog(ptr.address, "Node.Join", "default", "Info", "")

	var sucSucList [successorListSize]string
	node = GenerateWrapNode(sucAddr)
	err = node.GetSuccessorList(&sucSucList)

	if err != nil {
		createLog(ptr.address, "Node.Join", "RemoteCall.GetSuccessorList", "Error", err.Error())
		return false
	}
	ptr.nodeLock.Lock()
	ptr.predecessor = ""
	ptr.successorList[0] = sucAddr
	ptr.fingerTable[0] = sucAddr
	for i := 1; i < successorListSize; i++ {
		ptr.successorList[i] = sucSucList[i-1]
	}
	ptr.nodeLock.Unlock()

	node = GenerateWrapNode(sucAddr)

	ptr.dataManagerPtr.storageLock.Lock()
	err = node.InheritData(ptr.address, &ptr.dataManagerPtr.storage)
	ptr.dataManagerPtr.storageLock.Unlock()
	if err != nil {
		createLog(ptr.address, "Node.Join", "RemoteCall.inheritData", "Error", err.Error())
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
	err := ptr.server.ShutDown()
	if err != nil {
		createLog(ptr.address, "Node.Quit", "network.ShutDown", "Error", err.Error())
	}
	ptr.nodeLock.Lock()
	ptr.readyOnline = false
	ptr.nodeLock.Unlock()

	var sucAddr string
	err = ptr.findFirstOnlineSuccessor(&sucAddr)
	node := GenerateWrapNode(sucAddr)
	err = node.FixPredecessor()
	if err != nil {
		createLog(ptr.address, "Node.Quit", "RemoteCall.FixPredecessor", "Error", err.Error())
	}
	err = node.Stabilize()
	if err != nil {
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
	err := ptr.server.ShutDown()
	if err != nil {
		createLog(ptr.address, "Node.ForceQuit", "network.ShutDown", "Error", err.Error())
	}
	ptr.nodeLock.Lock()
	ptr.readyOnline = false
	ptr.nodeLock.Unlock()

	ptr.clear()
	createLog(ptr.address, "Node.ForceQuit", "default", "Info", "")
}

// Ping /* Check whether the node represented by the IP address is in the network. */
func (ptr *Node) Ping(addr string) bool {
	return IfOnline(addr)
}

// Put /* Put a key-value pair into the network (if KEY is already in the network, cover it), or
func (ptr *Node) Put(key string, value string) bool {
	if !ptr.readyOnline {
		createLog(ptr.address, "Node.Put", "default", "Error", "self off line")
		return false
	}
	var targetAddr string
	err := ptr.findSuccessor(CalHash(key), &targetAddr)
	if err != nil {
		createLog(ptr.address, "Node.Put", "Node.findSuccessor", "Error", err.Error())
		return false
	}

	node := GenerateWrapNode(targetAddr)
	err = node.StoreData(key, value)

	if err != nil {
		createLog(ptr.address, "Node.Put", "RemoteCall.StoreData", "Error", err.Error())
		return false
	}
	createLog(ptr.address, "Node.Put", "default", "Info", "")
	return true
} /* Return "true" if success, "false" otherwise. */

func (ptr *Node) Get(key string) (bool, string) {
	createLog(ptr.address, "Node.Get", "default", "Info", "enter Node.Get")
	if !ptr.readyOnline {
		createLog(ptr.address, "Node.Get", "default", "Error", "self off line")
		return false, ""
	}
	var value string
	var targetAddr string
	err := ptr.findSuccessor(CalHash(key), &targetAddr)
	if err != nil {
		createLog(ptr.address, "Node.Get", "Node.findSuccessor", "Error", err.Error())
		return false, ""
	}
	node := GenerateWrapNode(targetAddr)
	err = node.GetData(key, &value)
	if err != nil {
		createLog(ptr.address, "Node.Get", "Node.GetData", "Error", err.Error())
		return false, ""
	}
	createLog(ptr.address, "Node.Get", "default", "Info", "")
	return true, value
} /* Return "true" and the value if success, "false" otherwise. */

func (ptr *Node) Delete(key string) bool {
	if !ptr.readyOnline {
		createLog(ptr.address, "Node.Delete", "default", "Error", "self off line")
		return false
	}
	var targetAddr string
	err := ptr.findSuccessor(CalHash(key), &targetAddr)
	if err != nil {
		createLog(ptr.address, "Node.Delete", "Node.findSuccessor", "Error", err.Error())
		return false
	}

	node := GenerateWrapNode(targetAddr)
	err = node.DeleteData(key)

	if err != nil {
		createLog(ptr.address, "Node.Delete", "Node.DeleteData", "Error", err.Error())
		return false
	}
	createLog(ptr.address, "Node.Delete", "default", "Info", "")
	return true
} /* Remove the key-value pair represented by KEY from the network. */
/* Return "true" if remove successfully, "false" otherwise. */

func (ptr *Node) DataSize() int {
	return len(ptr.dataManagerPtr.storage)
}

//--------------------private func----------------------
func createLog(addr string, callFuncName string, funcName string, logType string, info string) {
	// if logType == "Info" {
	// 	log.Infof("[%s] <%s> Call <%s> Succeed %s\n", addr, callFuncName, funcName, info)
	// } else if logType == "Warning" {
	// 	log.Warningf("[%s] <%s> Call <%s> Warning %s\n", addr, callFuncName, funcName, info)
	// } else if logType == "Error" {
	// 	log.Errorf("[%s] <%s> Call <%s> Error %s\n", addr, callFuncName, funcName, info)
	// }
}

func (ptr *Node) findSuccessor(targetId *big.Int, result *string) error {
	var sucAddr string
	_ = ptr.findFirstOnlineSuccessor(&sucAddr)
	sucId := CalHash(sucAddr)

	if targetId.Cmp(sucId) == 0 || In(targetId, ptr.Id, sucId) {
		*result = sucAddr
		return nil
	}
	createLog(ptr.address, "node.findSuccessor", "default", "Info", "findSuc last")
	node := GenerateWrapNode(ptr.closestPrecedingFinger(targetId))
	return node.FindSuccessor(targetId, result)
}

func (ptr *Node) closestPrecedingFinger(targetId *big.Int) string {
	for i := hashBitsSize - 1; i >= 0; i-- {
		if ptr.fingerTable[i] == "" || !IfOnline(ptr.fingerTable[i]) {
			continue
		}
		if In(CalHash(ptr.fingerTable[i]), ptr.Id, targetId) {
			return ptr.fingerTable[i]
		}
	}
	var preAddr string
	err := ptr.findFirstOnlineSuccessor(&preAddr)
	if err != nil {
		return ""
	}
	return preAddr
}

func (ptr *Node) findFirstOnlineSuccessor(result *string) error {
	createLog(ptr.address, "findFirstOnlineSuccessor", "default", "Info", "into find first online successor")

	for i := 0; i < successorListSize; i++ {
		if IfOnline(ptr.successorList[i]) {
			*result = ptr.successorList[i]
			return nil
		}
		createLog(ptr.address, "findFirstOnlineSuccessor", "default", "Waring", "didn't find online")
	}
	return errors.New("list Break")
}

func (ptr *Node) clear() {
	ptr.quitSignal = make(chan bool, 2)
	ptr.dataManagerPtr.clear()
	ptr.nodeLock.Lock()
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
	var sucPreAddr, newSucAddr string
	err := ptr.findFirstOnlineSuccessor(&newSucAddr)

	node := GenerateWrapNode(newSucAddr)
	err = node.GetPredecessor(&sucPreAddr)

	if err != nil {
		return
	}
	if sucPreAddr != "" && In(CalHash(sucPreAddr), ptr.Id, CalHash(newSucAddr)) {
		newSucAddr = sucPreAddr
	}

	var sucSucList [successorListSize]string
	node = GenerateWrapNode(newSucAddr)
	err = node.GetSuccessorList(&sucSucList)

	if err != nil {
		return
	}
	ptr.nodeLock.Lock()
	ptr.successorList[0] = newSucAddr
	ptr.fingerTable[0] = newSucAddr
	for i := 1; i < successorListSize; i++ {
		ptr.successorList[i] = sucSucList[i-1]
	}
	ptr.nodeLock.Unlock()

	node = GenerateWrapNode(newSucAddr)
	err = node.Notify(ptr.address)

	if err != nil {
		return
	}
}

func (ptr *Node) notify(addr string) error {
	if ptr.predecessor == addr {
		return nil
	}

	if ptr.predecessor == "" || In(CalHash(addr), CalHash(ptr.predecessor), ptr.Id) {
		ptr.nodeLock.Lock()
		ptr.predecessor = addr
		ptr.nodeLock.Unlock()
		node := GenerateWrapNode(addr)
		_ = node.GenerateBackup(&ptr.dataManagerPtr.preBackUp)
	}
	return nil
}

func (ptr *Node) fixPredecessor() error {
	if ptr.predecessor != "" && !IfOnline(ptr.predecessor) {
		ptr.nodeLock.Lock()
		ptr.predecessor = ""
		ptr.nodeLock.Unlock()
		_ = ptr.dataManagerPtr.applyBackup()
	}
	return nil
}

func (ptr *Node) fixFinger() {
	var sucAddr string
	err := ptr.findSuccessor(PlusTwoPower(ptr.Id, ptr.next), &sucAddr)
	if err != nil {
		createLog(ptr.address, "fixFinger", "findSuccessor", "Error", err.Error())
		return
	}
	ptr.nodeLock.Lock()
	ptr.fingerTable[0] = ptr.successorList[0]
	ptr.fingerTable[ptr.next] = sucAddr
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
			_ = ptr.fixPredecessor()
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
