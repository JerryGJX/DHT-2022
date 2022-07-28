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

	for {
		var para1, para2, para3, para4 string = "", "", "", ""
		fmt.Scanln(&para1, &para2, &para3, &para4)
		if para1 == "join" {
			ok := self.node.Join(para2)
			if ok {
				fmt.Println("Join ", para2, " Successfully!")
			} else {
				fmt.Println("Fail to Join ", para2)
			}
			continue
		}
		if para1 == "create" {
			self.node.Create()
			fmt.Println("Create new network in ", self.addr)
			continue
		}
		if para1 == "upload" {
			err := Lauch(para2, para3, &self.node)
			if err != nil {
				fmt.Println("Fail to upload ", para2)
			}
			continue
		}
		if para1 == "download" {
			err := download(para3, para4, &self.node)
			if err != nil {
				fmt.Println("Fail to download ", err)
			}
			continue
		}
		if para1 == "quit" {
			self.node.Quit()
			fmt.Println(self.addr, " Node Quit")
			continue
		}
		if para1 == "run" {
			self.node.Run()
			fmt.Println(self.addr, "Run successfully")
			continue
		}
		if para1 == "help" {
			fmt.Println("<--------------------------------------------------------------------------------------------------------------------------->")
			fmt.Println("|   join [IP address]              # Join the network composed of node in IP address                                         |")
			fmt.Println("|   create                         # Create a new network                                                                    |")
			fmt.Println("|   upload [path1] [path2]         # Upload file in path1 and generate .torrent in path2(Default is the current directory)   |")
			fmt.Println("|   download -t [path1] [path2]    # Download file by .torrent in path1 into path2(Default is current directory)             |")
			fmt.Println("|   download -m [value] [path]     # Download file by magnet into path(Default)                                              |")
			fmt.Println("|   quit                           # Quit the network                                                                        |")
			fmt.Println("<--------------------------------------------------------------------------------------------------------------------------->")
			continue
		}
		fmt.Println("Unknown instruction :(")
	}

}
