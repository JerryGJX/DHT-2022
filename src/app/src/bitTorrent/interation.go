package bitTorrent

import (
	"fmt"
)

var self Client
var port int
var bootstrapAddr string

func Welcome() {
	hiBlue.Println("Hello, this is a naive bitTorrent, Welcome.")

	hiBlue.Println("* Please Input your port and bootstrap address")
	fmt.Scanln(&port)
	self.Init(port)

	self.node.Run()

	blue.Println("type \"help\" for more info")

	for {
		var (
			op       = ""
			element1 = ""
			element2 = ""
			element3 = ""
		)
		fmt.Scanln(&op, &element1, &element2, &element3)
		if op == "join" {
			flag := self.node.Join(element2)
			if flag {
				green.Println("Join succeed")
			} else {
				red.Println("Join failed")
			}
			continue
		}
		if op == "create" {
			self.node.Create()
			green.Println("Create network in ", self.addr, "succeed")
			continue
		}

		if op == "upload" {
			self.Upload(element1, element2)
			green.Println("Upload succeed")
			continue
		}

		if op == "download" {
			if element1 == "-t" {
				self.DownLoadByMagnet(element2, element3)
			} else if element1 == "-m" {
				self.DownLoadByTorrent(element2, element3)
			} else {
				red.Println("either -t or -m is needed")
			}
			green.Println("download succeed")
			continue
		}

		if op == "quit" {
			self.Quit()
			green.Println("quit succeed, bye (:")
			continue
		}

		if op == "help" {
			yellow.Println("<----------------------------------------------------------------------------------------->")
			yellow.Println("   join        [address]                  # Join the network                               ")
			yellow.Println("   create                                 # Create a new network                           ")
			yellow.Println("   upload      [file path] [torrent]      # Upload file and generate .torrent              ")
			yellow.Println("   download -t [torrent]   [file path]    # Download file by .torrent                      ")
			yellow.Println("   download -m [magnet]    [file path]    # Download file by magnet                        ")
			yellow.Println("   quit                                   # Quit the network                               ")
			yellow.Println("<----------------------------------------------------------------------------------------->")
			continue
		}

		red.Println("invalid operation, please try again")

	}

}
