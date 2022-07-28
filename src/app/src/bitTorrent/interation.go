package bitTorrent

import (
	"fmt"
	"time"
)

var self Peer
var port int
var bootstrapAddr string

func Welcome() {
	hiBlue.Println("Hello, this is a naive bitTorrent, Welcome.")

	hiBlue.Println("* Please Input your port and bootstrap address")
	fmt.Scanln(&port, &bootstrapAddr)
	self.Login(port, bootstrapAddr)

	time.Sleep(AfterLoginSleep)

	
}
