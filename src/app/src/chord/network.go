package chord

import (
	"errors"
	//log "github.com/sirupsen/logrus"
	"net"
	"net/rpc"
	"time"
)

type network struct {
	serv    *rpc.Server
	lis     net.Listener
	nodePtr *RpcNode
}

const tryTime int = 5

func (n *network) Init(address string, ptr *Node) error {
	n.serv = rpc.NewServer()
	n.nodePtr = new(RpcNode)
	n.nodePtr.node = ptr
	//注册rpc服务
	err := n.serv.Register(n.nodePtr)
	if err != nil {
		createLog(n.nodePtr.node.address, "network.Init", "rpc.Register", "Error", err.Error())
		return err
	}
	n.lis, err = net.Listen("tcp", address)
	if err != nil {
		createLog(n.nodePtr.node.address, "network.Init", "rpc.Listen", "Error", err.Error())
		return err
	}
	createLog(n.nodePtr.node.address, "network.Init", "default", "Info", "")
	go WrappedAccept(n.serv, n.lis, n.nodePtr.node)
	return nil
}

func (n *network) ShutDown() error {
	n.nodePtr.node.quitSignal <- true
	err := n.lis.Close()
	if err != nil {
		createLog(n.nodePtr.node.address, "network.ShutDown", "Listener.Close", "Error", err.Error())
		return err
	}
	createLog(n.nodePtr.node.address, "network.ShutDown", "default", "Info", "")
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

func IfOnline(address string) bool {
	client, err := GenerateClient(address)
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

func RemoteCall(addr string, serviceMethod string, args interface{}, reply interface{}) error {
	if addr == "" {
		return errors.New("null address for RemoteCall")
	}

	client, err := GenerateClient(addr)
	if err != nil {
		createLog(addr, "network.RemoteCall", "network.GenerateClient", "Error", err.Error())
		return err
	}
	createLog(addr, "network.RemoteCall", "network.GenerateClient", "Info", "after GenerateClient")
	if client != nil {
		defer client.Close()
	}
	err2 := client.Call(serviceMethod, args, reply)
	//createLog(addr, "network.RemoteCall", "client.Call", "Info", "after call")
	if err2 != nil {
		createLog(addr, "network.RemoteCall", "client.Call", "Error", err2.Error())
	}
	return err2
}

func WrappedAccept(server *rpc.Server, lis net.Listener, ptr *Node) {
	for {
		conn, err := lis.Accept()
		select {
		case <-ptr.quitSignal:
			return
		default:
			if err != nil {
				createLog(ptr.address, "network.WrappedAccept", "listener.Accept", "Error", err.Error())
				return
			}
			go server.ServeConn(conn)
		}
	}
}
