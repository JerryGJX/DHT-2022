package Kademlia

import (
	"fmt"

	"sync/atomic"
	"time"
)

type Node struct {
	station   *network
	isRunning bool
	data      Datamanager
	table     RoutingTable
	addr      AddrType
}

func (ptr *Node) clear() {
	ptr.isRunning = false
	ptr.table.InitRoutingTable(ptr.addr)
	ptr.data.init()
}

func (ptr *Node) Init(port int) {
	ptr.addr = GenerateAddr(fmt.Sprintf("%s:%d", localAddress, port))
	ptr.clear()
}

func (ptr *Node) Run() {
	ptr.station = new(network)
	err := ptr.station.Init(ptr.addr.Ip, ptr)
	if err != nil {
		createLog(ptr.addr.Ip, "Node.Run", "network.Init", "Error", err.Error())
		return
	}
	createLog(ptr.addr.Ip, "Node.Run", "default", "Info", "")
	ptr.isRunning = true
	ptr.Background()
}

func (ptr *Node) Join(address string) bool {
	createLog(ptr.addr.Ip, "Node.Join", "default", "Info", "get in")

	tmp := GenerateAddr(address)
	if isOnline := IfOnline(address); !isOnline {

		// println("fuck")

		createLog(ptr.addr.Ip, "Node.Join", "network.IfOnline", "Warning", address+" is off line")
		return false
	}

	// println("fuck")

	createLog(ptr.addr.Ip, "Node.Join", "default", "Waring", "Before Update")

	ptr.table.Update(&tmp)
	ptr.FindClosestNode(ptr.addr)
	createLog(ptr.addr.Ip, "Node.Join", "default", "Info", "")
	return true
}

func (ptr *Node) Quit() {
	// println("fuck1")
	_ = ptr.station.ShutDown()
	// println("fuck2")
	ptr.clear()
	// println("fuck3")
	createLog(ptr.addr.Ip, "Node.Quit", "default", "Info", "")
}

func (ptr *Node) Ping(requester string) bool {
	return true
}

func (ptr *Node) FindClosestNode(target AddrType) []ClosestListNode {
	resultList := make([]ClosestListNode, 0, K*2)
	pendingList := ptr.table.FindClosest(target.Id, K)
	inRun := new(int32)
	*inRun = 0
	visit := make(map[string]bool)
	visit[ptr.addr.Ip] = true
	index := 0
	ch := make(chan FindNodeRep, alpha+3)
	for index < len(pendingList) && *inRun < alpha {
		tmpReplier := pendingList[index].Addr
		if _, ok := visit[tmpReplier.Ip]; !ok {
			visit[tmpReplier.Ip] = true
			atomic.AddInt32(inRun, 1)
			go func(Replier *AddrType, channel chan FindNodeRep) {
				var response FindNodeRep
				node := GenerateWrapNode(ptr, Replier.Ip)
				err := node.FindNode(&FindNodeArg{ptr.addr, target.Id}, &response)

				if err != nil {
					atomic.AddInt32(inRun, -1)
					createLog(ptr.addr.Ip, "Node.FindClosestNode", "FindNode", "Error", err.Error())
					return
				}
				channel <- response
				return
			}(&tmpReplier, ch)
		}
		index++
	}
	for index < len(pendingList) || *inRun > 0 {
		if *inRun > 0 {
			select {
			case response := <-ch:
				atomic.AddInt32(inRun, -1)
				resultList = append(resultList, ClosestListNode{Xor(response.RepAddr.Id, target.Id), response.RepAddr})
				for _, value := range response.Content {
					pendingList = append(pendingList, value)
				}
				Sort(&pendingList)
			case <-time.After(WaitTime):
				createLog(ptr.addr.Ip, "Node.FindClosestNode", "default", "Waring", "Avoid Blocking...")
			}
		}
		for index < len(pendingList) && *inRun < alpha {
			tmpReplier := pendingList[index].Addr
			if _, ok := visit[tmpReplier.Ip]; !ok {
				visit[tmpReplier.Ip] = true
				atomic.AddInt32(inRun, 1)
				go func(Replier *AddrType, channel chan FindNodeRep) {
					var response FindNodeRep
					node := GenerateWrapNode(ptr, Replier.Ip)
					err := node.FindNode(&FindNodeArg{ptr.addr, target.Id}, &response)
					if err != nil {
						atomic.AddInt32(inRun, -1)
						createLog(ptr.addr.Ip, "Node.FindClosestNode", "default", "Error", err.Error())
						return
					}
					channel <- response
					return
				}(&tmpReplier, ch)
			}
			index++
		}
	}
	Sort(&resultList)
	if len(resultList) > K {
		resultList = resultList[:K]
	}
	return resultList
}

func (ptr *Node) Refresh() {
	lastRefreshTime := ptr.table.refreshTimeSet[ptr.table.refreshIndex]
	if !lastRefreshTime.Add(refreshTime).After(time.Now()) {
		tmpAddr := GenerateAddr(ptr.addr.Ip)
		ptr.FindClosestNode(tmpAddr)
		ptr.table.refreshTimeSet[ptr.table.refreshIndex] = time.Now()
	}
	ptr.table.refreshIndex = (ptr.table.refreshIndex + 1) % (IDlength)
}

func (ptr *Node) Put(key string, value string) bool {
	request := StoreArg{key, value, root, ptr.addr}
	ptr.data.store(request)
	request.priType = publisher
	ptr.RangePut(request)
	createLog(ptr.addr.Ip, "Node.Put", "default", "Info", "")
	return true
}

func (ptr *Node) RangePut(request StoreArg) {
	pendingList := ptr.FindClosestNode(GenerateAddr(request.Key))
	count := new(int32)
	*count = 0
	for index := 0; index < len(pendingList); {
		if *count < alpha {
			target := pendingList[index].Addr
			index++
			atomic.AddInt32(count, 1)
			go func(input StoreArg, targetNode *AddrType) {
				node := GenerateWrapNode(ptr, targetNode.Ip)
				err := node.Store(&input)
				if err != nil {
					createLog(ptr.addr.Ip, "Node.RangePut", "default", "Error", "fail to put"+err.Error())

				}
				atomic.AddInt32(count, -1)
			}(request, &target)
		} else {
			time.Sleep(SleepTime)
		}
	}
}

func (ptr *Node) Get(key string) (bool, string) {
	ifFind := false
	reply := ""
	requestInfo := FindValueArg{key, ptr.addr}
	resultList := make([]AddrType, 0, K*2)
	pendingList := ptr.table.FindClosest(CalHash(key), K)
	inRun := new(int32)
	*inRun = 0
	visit := make(map[string]bool)
	visit[ptr.addr.Ip] = true
	index := 0
	ch := make(chan FindValueRep, alpha+3)
	for index < len(pendingList) && *inRun < alpha {
		tmpReplier := pendingList[index].Addr
		if _, ok := visit[tmpReplier.Ip]; !ok {
			visit[tmpReplier.Ip] = true
			atomic.AddInt32(inRun, 1)
			go func(Replier *AddrType, channel chan FindValueRep) {
				var response FindValueRep
				node := GenerateWrapNode(ptr, Replier.Ip)
				err := node.FindValue(&requestInfo, &response)
				if err != nil {
					atomic.AddInt32(inRun, -1)
					return
				}
				channel <- response
				return
			}(&tmpReplier, ch)
		}
		index++
	}
	for (index < len(pendingList) || *inRun > 0) && !ifFind {
		if *inRun > 0 {
			select {
			case response := <-ch:
				atomic.AddInt32(inRun, -1)
				if response.IfFind {
					ifFind = true
					reply = response.Value
					break
				}
				resultList = append(resultList, response.RepAddr)
				for _, value := range response.Content {
					pendingList = append(pendingList, value)
				}
				Sort(&pendingList)
			case <-time.After(WaitTime):
			}
			if ifFind {
				break
			}
		}
		for index < len(pendingList) && *inRun < alpha && !ifFind {
			tmpReplier := pendingList[index].Addr
			if _, ok := visit[tmpReplier.Ip]; !ok {
				visit[tmpReplier.Ip] = true
				atomic.AddInt32(inRun, 1)
				go func(Replier *AddrType, channel chan FindValueRep) {
					var response FindValueRep
					node := GenerateWrapNode(ptr, Replier.Ip)
					err := node.FindValue(&requestInfo, &response)
					if err != nil {
						atomic.AddInt32(inRun, -1)
						return
					}
					channel <- response
					return
				}(&tmpReplier, ch)
			}
			index++
		}
	}
	if !ifFind {
		return false, ""
	} else {
		StoreInfo := StoreArg{key, reply, duplicater, ptr.addr}
		count := new(int32)
		*count = 0
		for i := 0; i < len(resultList); {
			if *count < alpha {
				target := resultList[i]
				i++
				atomic.AddInt32(count, 1)
				go func(input StoreArg, targetNode *AddrType) {
					node := GenerateWrapNode(ptr, targetNode.Ip)
					err := node.Store(&input)
					if err != nil {
						createLog(ptr.addr.Ip, "Node.Get", "default", "Error", "fail to Get"+err.Error())

					}
					atomic.AddInt32(count, -1)
				}(StoreInfo, &target)
			} else {
				time.Sleep(SleepTime)
			}
		}
		return true, reply
	}
}

func (ptr *Node) Republic() {
	pendingList := ptr.data.republish()
	for k, v := range pendingList {
		request := StoreArg{k, v, publisher, ptr.addr}
		ptr.RangePut(request)
	}
}

func (ptr *Node) Duplicate() {
	pendingList := ptr.data.duplicate()
	for k, v := range pendingList {
		request := StoreArg{k, v, duplicater, ptr.addr}
		ptr.RangePut(request)
	}
}

func (ptr *Node) Expire() {
	ptr.data.clearExpire()
}

func (ptr *Node) Background() {
	go func() {
		for ptr.isRunning {
			ptr.Refresh()
			time.Sleep(backgroundLow)
		}
	}()
	go func() {
		for ptr.isRunning {
			ptr.Duplicate()
			time.Sleep(backgroundHigh)
		}
	}()
	go func() {
		for ptr.isRunning {
			ptr.Expire()
			time.Sleep(backgroundHigh)
		}
	}()
	go func() {
		for ptr.isRunning {
			ptr.Republic()
			time.Sleep(backgroundHigh)
		}
	}()
}

// Create unused function
func (ptr *Node) Create() {

}

func (ptr *Node) ForceQuit() {
	_ = ptr.station.ShutDown()
	ptr.clear()
}

func (ptr *Node) Delete(key string) bool {
	return true
}
