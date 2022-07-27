package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)


var (
	ourNode dhtNode
	ourIp string
)



func init() {
	f, _ := os.Create("out.log")
	log.SetOutput(f)
	
	fmt.Println("Please type your IP to quick start :)")
	fmt.Scanln(&ourIp)
	fmt.Println("IP is set to ", ourIp, ", type help to get command. ")
	ourNode = NewNode(ourIp)
	ourNode.Run()
}