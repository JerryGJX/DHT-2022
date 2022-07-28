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
	base = big.NewInt(2)
	Mod = new(big.Int).Exp(base, big.NewInt(160), nil)
	CutTime = 100 * time.Millisecond
	WaitTime = 250 * time.Millisecond
}

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

//----------------------------

type Pair struct {
	Key   string
	Value string
}

/*(a,b)*/

func In(target, low, high *big.Int) bool {
	if val := low.Cmp(high); val < 0 {
		return low.Cmp(target) < 0 && high.Cmp(target) > 0
	} else if val == 0 {
		return low.Cmp(target) != 0
	}
	//low>high
	return low.Cmp(target) < 0 || high.Cmp(target) > 0
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
