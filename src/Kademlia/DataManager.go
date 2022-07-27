package Kademlia

import (
	"github.com/sasha-s/go-deadlock"
	log "github.com/sirupsen/logrus"
	"time"
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

type database struct {
	rwLock        deadlock.RWMutex
	storage       map[string]string
	expireTime    map[string]time.Time
	republishTime map[string]time.Time
	duplicateTime map[string]time.Time
	privilege     map[string]int
}

func (ptr *database) init() {
	ptr.storage = make(map[string]string)
	ptr.expireTime = make(map[string]time.Time)
	ptr.duplicateTime = make(map[string]time.Time)
	ptr.republishTime = make(map[string]time.Time)
	ptr.privilege = make(map[string]int)
}

func (ptr *database) store(request StoreArg) {
	ptr.rwLock.Lock()
	defer ptr.rwLock.Unlock()
	if _, ok := ptr.storage[request.Key]; !ok {
		requestPri := request.RequesterPri
		ptr.privilege[request.Key] = requestPri + 1
		ptr.storage[request.Key] = request.Value
		if requestPri == root {
			ptr.republishTime[request.Key] = time.Now().Add(tRepublish)
			return
		}
		if requestPri == publisher {
			ptr.duplicateTime[request.Key] = time.Now().Add(tDuplicate)
			ptr.expireTime[request.Key] = time.Now().Add(expireTimeInterval2)
			return
		}
		if requestPri == duplicater {
			ptr.expireTime[request.Key] = time.Now().Add(expireTimeInterval3)
			return
		}
		log.Errorf("<database store> Wrong privilege1")
	} else {
		originPri := ptr.privilege[request.Key]
		requestPri := request.RequesterPri
		if originPri == publisher || requestPri >= originPri {
			return
		}
		// duplicater->publisher || common->publisher || common->duplicater
		if requestPri == root {
			ptr.privilege[request.Key] = publisher
			ptr.storage[request.Key] = request.Value
			ptr.republishTime[request.Key] = time.Now().Add(tRepublish)
			delete(ptr.expireTime, request.Key)
			delete(ptr.duplicateTime, request.Key)
			return
		}
		if requestPri == publisher {
			ptr.privilege[request.Key] = duplicater
			ptr.storage[request.Key] = request.Value
			ptr.expireTime[request.Key] = time.Now().Add(expireTimeInterval2)
			ptr.duplicateTime[request.Key] = time.Now().Add(tDuplicate)
			delete(ptr.republishTime, request.Key)
			return
		}
		if requestPri == duplicater {
			ptr.privilege[request.Key] = cacher
			ptr.storage[request.Key] = request.Value
			ptr.expireTime[request.Key] = time.Now().Add(expireTimeInterval3)
			delete(ptr.duplicateTime, request.Key)
			delete(ptr.republishTime, request.Key)
			return
		}
		log.Errorf("<database store> Wrong privilege2")
	}
}

func (ptr *database) clearExpire() {
	tmp := make(map[string]bool)
	ptr.rwLock.RLock()
	for key, value := range ptr.expireTime {
		if !value.After(time.Now()) {
			tmp[key] = true
		}
	}
	ptr.rwLock.RUnlock()
	ptr.rwLock.Lock()
	for key:= range tmp {
		delete(ptr.storage, key)
		delete(ptr.expireTime, key)
		delete(ptr.duplicateTime, key)
		delete(ptr.republishTime, key)
		delete(ptr.privilege, key)
	}
	ptr.rwLock.Unlock()
}

func (ptr *database) duplicate() (result map[string]string) {
	result = make(map[string]string)
	ptr.rwLock.RLock()
	for key, value := range ptr.duplicateTime {
		if !value.After(time.Now()) {
			result[key] = ptr.storage[key]
		}
	}
	ptr.rwLock.RUnlock()
	ptr.rwLock.Lock()
	for key:= range result {
		ptr.duplicateTime[key] = time.Now().Add(tDuplicate)
	}
	ptr.rwLock.Unlock()
	return
}

func (ptr *database) republic() (result map[string]string) {
	result = make(map[string]string)
	ptr.rwLock.RLock()
	for k, v := range ptr.republishTime {
		if !v.After(time.Now()) {
			result[k] = ptr.storage[k]
		}
	}
	ptr.rwLock.RUnlock()
	ptr.rwLock.Lock()
	for k:= range result {
		ptr.republishTime[k] = time.Now().Add(tRepublish)
	}
	ptr.rwLock.Unlock()
	return
}

func (ptr *database) get(key string) (bool, string) {
	ptr.rwLock.Lock()
	defer ptr.rwLock.Unlock()
	if v, ok := ptr.storage[key]; ok {
		if _, ok2 := ptr.expireTime[key]; ok2 {
			if !ptr.expireTime[key].After(time.Now().Add(expireTimeInterval3)) {
				ptr.expireTime[key] = time.Now().Add(expireTimeInterval3)
			}
		}
		return true, v
	} else {
		return false, ""
	}
}
