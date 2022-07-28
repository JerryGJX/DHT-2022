package Kademlia

import (
	"crypto/sha1"
	"math/big"
	"net"
	"sort"

	log "github.com/sirupsen/logrus"
)

func GetLocalIP() string {
	var localaddress string
	ifaces, err := net.Interfaces()
	if err != nil {
		panic("init: failed to find network interfaces")
	}
	for _, elt := range ifaces {
		if elt.Flags&net.FlagLoopback == 0 && elt.Flags&net.FlagUp != 0 {
			addrs, err := elt.Addrs()
			if err != nil {
				panic("init: failed to get addresses for network interface")
			}

			for _, addr := range addrs {
				ipnet, ok := addr.(*net.IPNet)
				if ok {
					if ip4 := ipnet.IP.To4(); len(ip4) == net.IPv4len {
						localaddress = ip4.String()
						break
					}
				}
			}
		}
	}
	if localaddress == "" {
		panic("init: failed to find non-loopback interface with valid address on this node")
	}
	return localaddress
}

func createLog(addr string, callFuncName string, funcName string, logType string, info string) {
	if logType == "Info" {
		log.Infof("[%s] <%s> Call <%s> Succeed %s\n", addr, callFuncName, funcName, info)
	} else if logType == "Warning" {
		log.Warningf("[%s] <%s> Call <%s> Warning %s\n", addr, callFuncName, funcName, info)
	} else if logType == "Error" {
		log.Errorf("[%s] <%s> Call <%s> Error %s\n", addr, callFuncName, funcName, info)
	}
}

func Sort(dataSet *[]ClosestListNode) {
	sort.Slice(*dataSet, func(i, j int) bool {
		return (*dataSet)[i].Key.Cmp((*dataSet)[j].Key) < 0
	})
}

//----------------------------

type AddrType struct {
	Ip string
	Id *big.Int
}

func GenerateAddr(addr string) (result AddrType) {
	result.Ip = addr
	result.Id = CalHash(addr)
	return
}

func CalHash(str string) (result *big.Int) {
	h := sha1.New()
	h.Write([]byte(str))
	return (&big.Int{}).SetBytes(h.Sum(nil))
}

func CalNodeId(origin AddrType, adder int64) (result AddrType) {
	result.Id.Add(origin.Id, big.NewInt(adder))
	return result
}

func (node *AddrType) Equals(other AddrType) bool {
	return node.Id.Cmp(other.Id) == 0
}

func (node *AddrType) Copy(other AddrType) {
	node.Ip = other.Ip
	node.Id = CalHash(other.Ip)
}

func (node *AddrType) Less(other AddrType) bool {
	return node.Id.Cmp(other.Id) < 0
}

func Xor(node, other *big.Int) (result *big.Int) {
	result = new(big.Int)
	return result.Xor(node, other)
}

func PrefixLen(Id *big.Int) int {
	if Id.BitLen()-1 == -1 {
		return 0
	}
	return Id.BitLen() - 1

}
