package Kademlia

import (
	"time"

	"github.com/sasha-s/go-deadlock"
	log "github.com/sirupsen/logrus"
)

const (
	root       int = 0
	publisher  int = 1
	duplicater int = 2
	cacher     int = 3
)

const (
	tRepublish time.Duration = 5 * time.Hour
	tDuplicate time.Duration = 15 * time.Minute
)

type Datamanager struct {
	rwLock        deadlock.RWMutex
	storage       map[string]string
	expireTime    map[string]time.Time
	republishTime map[string]time.Time
	duplicateTime map[string]time.Time
	privilege     map[string]int
}

func (ptr *Datamanager) init() {
	ptr.storage = make(map[string]string)
	ptr.expireTime = make(map[string]time.Time)
	ptr.duplicateTime = make(map[string]time.Time)
	ptr.republishTime = make(map[string]time.Time)
	ptr.privilege = make(map[string]int)
}

func (ptr *Datamanager) store(info StoreArg) {
	ptr.rwLock.Lock()
	defer ptr.rwLock.Unlock()
	if _, flag := ptr.storage[info.Key]; !flag {
		reqPri := info.priType
		ptr.privilege[info.Key] = reqPri + 1
		ptr.storage[info.Key] = info.Value
		if reqPri == root {
			ptr.republishTime[info.Key] = time.Now().Add(tRepublish)
			return
		}
		if reqPri == publisher {
			ptr.duplicateTime[info.Key] = time.Now().Add(tDuplicate)
			ptr.expireTime[info.Key] = time.Now().Add(expireTimeHigh)
			return
		}
		if reqPri == duplicater {
			ptr.expireTime[info.Key] = time.Now().Add(expireTimeLow)
			return
		}
		log.Errorf("<database store> Wrong privilege1")
	} else {
		prePri := ptr.privilege[info.Key]
		reqPri := info.priType
		if prePri == publisher || reqPri >= prePri {
			return
		}
		if reqPri == root {
			ptr.privilege[info.Key] = publisher
			ptr.storage[info.Key] = info.Value
			ptr.republishTime[info.Key] = time.Now().Add(tRepublish)
			delete(ptr.expireTime, info.Key)
			delete(ptr.duplicateTime, info.Key)
			return
		}
		if reqPri == publisher {
			ptr.privilege[info.Key] = duplicater
			ptr.storage[info.Key] = info.Value
			ptr.expireTime[info.Key] = time.Now().Add(expireTimeHigh)
			ptr.duplicateTime[info.Key] = time.Now().Add(tDuplicate)
			delete(ptr.republishTime, info.Key)
			return
		}
		if reqPri == duplicater {
			ptr.privilege[info.Key] = cacher
			ptr.storage[info.Key] = info.Value
			ptr.expireTime[info.Key] = time.Now().Add(expireTimeLow)
			delete(ptr.duplicateTime, info.Key)
			delete(ptr.republishTime, info.Key)
			return
		}
	}
}

func (ptr *Datamanager) clearExpire() {
	tmp := make(map[string]bool)
	ptr.rwLock.RLock()
	for key, value := range ptr.expireTime {
		if !value.After(time.Now()) {
			tmp[key] = true
		}
	}
	ptr.rwLock.RUnlock()
	ptr.rwLock.Lock()
	for key := range tmp {
		delete(ptr.storage, key)
		delete(ptr.expireTime, key)
		delete(ptr.duplicateTime, key)
		delete(ptr.republishTime, key)
		delete(ptr.privilege, key)
	}
	ptr.rwLock.Unlock()
}

func (ptr *Datamanager) duplicate() (result map[string]string) {
	result = make(map[string]string)
	ptr.rwLock.RLock()
	for key, value := range ptr.duplicateTime {
		if !value.After(time.Now()) {
			result[key] = ptr.storage[key]
		}
	}
	ptr.rwLock.RUnlock()
	ptr.rwLock.Lock()
	for key := range result {
		ptr.duplicateTime[key] = time.Now().Add(tDuplicate)
	}
	ptr.rwLock.Unlock()
	return
}

func (ptr *Datamanager) republish() (result map[string]string) {
	result = make(map[string]string)
	ptr.rwLock.RLock()
	for k, v := range ptr.republishTime {
		if !v.After(time.Now()) {
			result[k] = ptr.storage[k]
		}
	}
	ptr.rwLock.RUnlock()
	ptr.rwLock.Lock()
	for k := range result {
		ptr.republishTime[k] = time.Now().Add(tRepublish)
	}
	ptr.rwLock.Unlock()
	return
}

func (ptr *Datamanager) get(key string) (bool, string) {
	ptr.rwLock.Lock()
	defer ptr.rwLock.Unlock()
	if value, flag := ptr.storage[key]; flag {
		if _, flag = ptr.expireTime[key]; flag {
			if !ptr.expireTime[key].After(time.Now().Add(expireTimeLow)) {
				ptr.expireTime[key] = time.Now().Add(expireTimeLow)
			}
		}
		return true, value
	} else {
		return false, ""
	}
}
