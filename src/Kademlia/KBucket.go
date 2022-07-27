package Kademlia

import (
	"container/list"
	"github.com/sasha-s/go-deadlock"
	"math/big"
	"time"
)

type RoutingTable struct {
	nodeAddr       AddrType
	rwLock         deadlock.RWMutex
	buckets        [IDlength]*list.List
	refreshIndex   int
	refreshTimeSet [IDlength]time.Time
}

func (ptr *RoutingTable) InitRoutingTable(nodeAddr AddrType) {
	ptr.rwLock.Lock()
	ptr.nodeAddr = nodeAddr
	for i := 0; i < IDlength; i++ {
		ptr.buckets[i] = list.New()
		ptr.refreshTimeSet[i] = time.Now()
	}
	ptr.refreshIndex = 0
	ptr.rwLock.Unlock()
}

// Update  used when replier node is called and requester call successfully
func (ptr *RoutingTable) Update(contact *AddrType) {
	createLog(ptr.nodeAddr.Ip,"RoutingTable.Update","PrefixLen1","Waring","after get bucket")

	//log.Infoln("<Update> Update ",contact.Address)
	ptr.rwLock.RLock()

	createLog(ptr.nodeAddr.Ip,"RoutingTable.Update","PrefixLen","Waring","after get bucket")
	
	bucket := ptr.buckets[PrefixLen(Xor(ptr.nodeAddr.Id, contact.Id))]
	
	
	target := bucket.Front()
	target = nil
	for i := bucket.Front(); ; i = i.Next() {
		if i == nil {
			target = nil
			break
		}
		if i.Value.(*AddrType).Equals(*contact) {
			target = i
			break
		}
	}
	ptr.rwLock.RUnlock()
	ptr.rwLock.Lock()
	if target != nil {
		bucket.MoveToBack(target)
	} else {
		if bucket.Len() < K {
			bucket.PushBack(contact)
		} else {
			tmp := bucket.Front()
			if !IfOnline(tmp.Value.(*AddrType).Ip) {
				bucket.Remove(tmp)
				bucket.PushBack(contact)
			} else {
				bucket.MoveToBack(tmp)
			}
		}
	}
	ptr.rwLock.Unlock()
}

func (ptr *RoutingTable) FindClosest(targetID *big.Int, count int) []ContactRecord {
	result := make([]ContactRecord, 0, count)
	index := PrefixLen(Xor(ptr.nodeAddr.Id, targetID))
	ptr.rwLock.RLock()
	if targetID.Cmp(ptr.nodeAddr.Id) == 0 {
		result = append(result, ContactRecord{Xor(targetID, targetID), GenerateAddr(ptr.nodeAddr.Ip)})
	}
	for i := ptr.buckets[index].Front(); i != nil && len(result) < count; i = i.Next() {
		contact := i.Value.(*AddrType)
		result = append(result, ContactRecord{Xor(targetID, contact.Id), *contact})
	}

	for i := 1; (index-i >= 0 || index+i < IDlength) && len(result) < count; i++ {
		if index-i >= 0 {
			for j := ptr.buckets[index-i].Front(); j != nil && len(result) < count; j = j.Next() {
				contact := j.Value.(*AddrType)
				result = append(result, ContactRecord{Xor(targetID, contact.Id), *contact})
			}
		}
		if index+i < IDlength {
			for j := ptr.buckets[index+i].Front(); j != nil && len(result) < count; j = j.Next() {
				contact := j.Value.(*AddrType)
				result = append(result, ContactRecord{Xor(targetID, contact.Id), *contact})
			}
		}
	}
	ptr.rwLock.RUnlock()
	return result
}
