package Kademlia

import (
	"math/big"
	"time"
)

const IDlength int = 160 // bytes

const K int = 20 // buckets size
const alpha int32 = 3
const tryTimes int = 4

var localAddress string = GetLocalIP()

const WaitTime time.Duration = 250 * time.Millisecond
const SleepTime time.Duration = 20 * time.Millisecond
const refreshTime time.Duration = 30 * time.Second
const expireTimeHigh time.Duration = 6 * time.Hour
const expireTimeLow time.Duration = 20 * time.Minute

const backgroundLow time.Duration = 5 * time.Second
const backgroundHigh time.Duration = 10 * time.Minute

type ClosestListNode struct {
	Key  *big.Int
	Addr AddrType
}
