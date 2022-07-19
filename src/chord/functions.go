package chord

import (
	"crypto/sha1"
	"math/big"
	"net"
	"time"
)

var (
	localAddr string
	Mod       *big.Int
	base      *big.Int
	CutTime   time.Duration
	WaitTime  time.Duration
)

func init() {
	localAddr = GetLocalIP()
	//localAddr = "127.0.0.1"
	base = big.NewInt(2)
	Mod = new(big.Int).Exp(base, big.NewInt(160), nil)
	CutTime = 200 * time.Millisecond
	WaitTime = 200 * time.Millisecond
}

func GetLocalIP() string {
	var localaddress string

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		panic("init: failed to find network interfaces")
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				localaddress = ipnet.IP.String()
				break
			}
		}
	}
	if localaddress == "" {
		panic("init: failed to find non-loopback interface with valid address on this node")
	}
	return localaddress
}

//----------------------------

type Pair struct {
	Key   string
	Value string
}

type Identifier struct {
	Value *big.Int
}

/*(a,b)*/
func (ID *Identifier) In(low, high *big.Int) bool {
	if val := low.Cmp(high); val < 0 {
		return low.Cmp(ID.Value) < 0 && high.Cmp(ID.Value) > 0
	} else if val == 0 {
		return low.Cmp(ID.Value) != 0
	}
	//low>high
	return low.Cmp(ID.Value) > 0 || high.Cmp(ID.Value) < 0
}

func (ID *Identifier) InLeftClose(low, high *big.Int) bool {
	return (low.Cmp(ID.Value) == 0 && low.Cmp(high) != 0) || ID.In(low, high)
}

func (ID *Identifier) InRightClose(low, high *big.Int) bool {
	if low.Cmp(high) != -1 {
		return false
	}
	return (high.Cmp(ID.Value) == 0 && low.Cmp(high) != 0) || ID.In(low, high)
}

func (ID *Identifier) Copy(rhs *Identifier) {
	ID = rhs
}

func GenerateId(str string) Identifier {
	var Id Identifier
	Id.Value = CalHash(str)
	return Id
}

func CalHash(str string) *big.Int {
	h := sha1.New()
	h.Write([]byte(str))
	return (&big.Int{}).SetBytes(h.Sum(nil))
}

// PlusTwoPower (raw+2^exp)%Mod
func PlusTwoPower(raw *big.Int, exp int) *big.Int {
	d := new(big.Int).Exp(base, big.NewInt(int64(exp)), nil)
	ans := new(big.Int).Add(raw, d)
	ans = new(big.Int).Mod(ans, Mod)
	return ans
}

//----------------------
//type Address struct {
//	Id   *Identifier
//	Addr string
//}

//func (ptr *Address) isNil() bool {
//	return ptr.Addr == ""
//}
//
//func (ptr *Address) clear() {
//	ptr.Addr = ""
//	ptr.Id.Value = nil
//}
//
//// Copy operator "="
//func (ptr *Address) Copy(rhs *Address) {
//	ptr.Addr = rhs.Addr
//	ptr.Id.Copy(rhs.Id)
//}
//
//func GenerateAddress(addr string) (reply Address) {
//	reply.Addr = addr
//	reply.Id = CalHash(addr)
//	return
//}
