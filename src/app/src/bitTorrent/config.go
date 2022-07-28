package bitTorrent

import (
	"fmt"
	"net"
	"time"

	"github.com/fatih/color"
)

var (
	green  = color.New(color.FgGreen)
	red    = color.New(color.FgRed)
	yellow = color.New(color.FgYellow)
	cyan   = color.New(color.FgCyan)
	blue   = color.New(color.FgBlue)
	hiBlue = color.New(color.FgHiBlue)
	magenta = color.New(color.FgHiMagenta)
)

const (
	PieceSize = 262144
	TimeWait = 3 * time.Second
	SHA1Len = 20
	WorkQueueBuffer = 128
	AfterLoginSleep = time.Second
	 AfterQuitSleep = time.Second

	 UploadTimeout = time.Second
	 DownloadTimeout = time.Second

	 RetryTime = 2

	 UploadInterval = 100 * time.Millisecond
	 DownloadInterval = 100 * time.Millisecond
	 DownloadWriteInterval = time.Second
	 UploadFileInterval = time.Second

)


func MakeMagnet(infoHash string) string {
	return "magnet:?xt=urn:btih:" + infoHash
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



func GenerateAddr (addr string,port int) string{
	return fmt.Sprintf("%s:%d",addr,port)
}