package Kademlia

import (
	"math/big"
	"time"
)

const IDlength int = 160 // bytes

const K int = 20 // buckets size
const alpha int32 = 3
const tryTimes int = 4
const localAddress string = "127.0.0.1"
const WaitTime time.Duration = 250 * time.Millisecond // with use of select
const SleepTime time.Duration = 20 * time.Millisecond // avoiding endless for loop
const refreshTimeInterval time.Duration = 30 * time.Second
const expireTimeInterval2 time.Duration = 6 * time.Hour
const expireTimeInterval3 time.Duration = 20 * time.Minute

//const republicTimeInterval time.Duration = 5 * time.Hour
//const duplicateTimeInterval time.Duration = 15 * time.Minute
const backgroundInterval1 time.Duration = 5 * time.Second
const backgroundInterval2 time.Duration = 10 * time.Minute

type ContactRecord struct {
	SortKey     big.Int
	ContactInfo AddrType
}
