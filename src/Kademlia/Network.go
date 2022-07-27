package Kademlia

import (
	"errors"
	//log "github.com/sirupsen/logrus"
	"net"
	"net/rpc"
	"time"
)

const tryTime int = 5

type network struct {
	serv       *rpc.Server
	lis        net.Listener
	nodePtr    *RpcNode
	quitSignal chan bool
}

func (n *network) Init(address string, ptr *Node) error {
	n.serv = rpc.NewServer()
	n.nodePtr = new(RpcNode)
	n.nodePtr.node = ptr
	//注册rpc服务
	err := n.serv.Register(n.nodePtr)
	if err != nil {
		createLog(n.nodePtr.node.addr.Ip, "network.Init", "rpc.Register", "Error", err.Error())
		return err
	}
	n.lis, err = net.Listen("tcp", address)
	if err != nil {
		createLog(n.nodePtr.node.addr.Ip, "network.Init", "rpc.Listen", "Error", err.Error())
		return err
	}
	createLog(n.nodePtr.node.addr.Ip, "network.Init", "default", "Info", "")
	go WrappedAccept(n.serv, n.lis, n.nodePtr.node)
	return nil
}

func (n *network) ShutDown() error {
	n.quitSignal <- true
	err := n.lis.Close()
	if err != nil {
		createLog(n.nodePtr.node.addr.Ip, "network.ShutDown", "Listener.Close", "Error", err.Error())
		return err
	}
	createLog(n.nodePtr.node.addr.Ip, "network.ShutDown", "default", "Info", "")
	return nil
}

func GenerateClient(address string) (*rpc.Client, error) {
	if address == "" {
		createLog(address, "network.GenerateClient", "default", "Error", "self off line")
		return nil, errors.New("<GetClient> IP address is nil")
	}
	var (
		err    error
		client *rpc.Client
	)
	ch := make(chan error)
	for i := 0; i < tryTime; i++ {
		go func() {
			client, err = rpc.Dial("tcp", address)
			ch <- err
		}()
		select {
		case <-ch:
			if err == nil {
				return client, nil
			} else {
				return nil, err
			}
		case <-time.After(WaitTime):
			err = errors.New("timeout")
			createLog(address, "network.GenerateClient", "rpc.Dial", "Error", err.Error())
		}
	}
	return nil, err
}

func PingAndCheckOnline(self *Node, addr *AddrType) bool {
	node := GenerateWrapNode(self, addr.Ip)
	err := node.Ping(self.addr)
	if err != nil {
		return false
	} else {
		return true
	}
}

func IfOnline(address string) bool {
// println("fuck1")
	client, err := GenerateClient(address)
	// defer println("fuck2")
	if err != nil {
		createLog(address, "network.IfOnline", "network.GenerateClient", "Error", err.Error())
		return false
	}
	if client != nil {
		defer client.Close()
	} else {
		return false
	}
	createLog(address, "network.IfOnline", "default", "Info", "")
	return true
}

func RemoteCall(self *Node, targetAddr *AddrType, serviceMethod string, args interface{}, reply interface{}) error {
	if targetAddr.Ip == "" {
		return errors.New("null address for RemoteCall")
	}

	client, err := GenerateClient(targetAddr.Ip)
	if err != nil {
		createLog(targetAddr.Ip, "network.RemoteCall", "network.GenerateClient", "Error", err.Error())
		return err
	}
	createLog(targetAddr.Ip, "network.RemoteCall", "network.GenerateClient", "Info", "after GenerateClient")
	if client != nil {
		self.table.Update(targetAddr)
		defer client.Close()
	}
	err2 := client.Call(serviceMethod, args, reply)
	//createLog(targetAddr, "network.RemoteCall", "client.Call", "Info", "after call")
	if err2 != nil {
		createLog(targetAddr.Ip, "network.RemoteCall", "client.Call   "+serviceMethod, "Error", err2.Error())
	}
	return err2
}

func WrappedAccept(server *rpc.Server, lis net.Listener, ptr *Node) {
	for {
		conn, err := lis.Accept()
		// println("fuck")
		select {
		case <-ptr.station.quitSignal:
			return
		default:
			if err != nil {
				createLog(ptr.addr.Ip, "network.WrappedAccept", "listener.Accept", "Error", err.Error())
				return
			}
			go server.ServeConn(conn)
		}
	}
}
