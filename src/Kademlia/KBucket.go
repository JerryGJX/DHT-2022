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

func (this *RoutingTable) InitRoutingTable(nodeAddr AddrType) {
	this.nodeAddr = nodeAddr
	this.rwLock.Lock()
	for i := 0; i < IDlength; i++ {
		this.buckets[i] = list.New()
		this.refreshTimeSet[i] = time.Now()
	}
	this.refreshIndex = 0
	this.rwLock.Unlock()
}

// Update  used when replier node is called and requester call successfully
func (this *RoutingTable) Update(contact *AddrType) {
	//log.Infoln("<Update> Update ",contact.Address)
	this.rwLock.RLock()
	bucket := this.buckets[PrefixLen(Xor(&(this.nodeAddr.Id), &(contact.Id)))]
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
	this.rwLock.RUnlock()
	this.rwLock.Lock()
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
	this.rwLock.Unlock()
}

func (this *RoutingTable) FindClosest(targetID big.Int, count int) []ContactRecord {
	result := make([]ContactRecord, 0, count)
	index := PrefixLen(Xor(&(this.nodeAddr.Id), &targetID))
	this.rwLock.RLock()
	if targetID.Cmp(&(this.nodeAddr.Id)) == 0 {
		result = append(result, ContactRecord{*Xor(&targetID, &targetID), GenerateAddr(this.nodeAddr.Ip)})
	}
	for i := this.buckets[index].Front(); i != nil && len(result) < count; i = i.Next() {
		contact := i.Value.(*AddrType)
		result = append(result, ContactRecord{*Xor(&targetID, &(contact.Id)), *contact})
	}
	for i := 1; (index-i >= 0 || index+i < IDlength*8) && len(result) < count; i++ {
		if index-i >= 0 {
			for j := this.buckets[index-i].Front(); j != nil && len(result) < count; j = j.Next() {
				contact := j.Value.(*AddrType)
				result = append(result, ContactRecord{*Xor(&targetID, &(contact.Id)), *contact})
			}
		}
		if index+i < IDlength {
			for j := this.buckets[index+i].Front(); j != nil && len(result) < count; j = j.Next() {
				contact := j.Value.(*AddrType)
				result = append(result, ContactRecord{*Xor(&targetID, &(contact.Id)), *contact})
			}
		}
	}
	this.rwLock.RUnlock()
	return result
}
