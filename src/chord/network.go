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
	nodePtr *WrapNode
}

const tryTime int = 5

func (n *network) Init(address string, ptr *Node) error {
	n.serv = rpc.NewServer()
	n.nodePtr = new(WrapNode)
	n.nodePtr.node = ptr
	//注册rpc服务
	err := n.serv.Register(n.nodePtr)
	if err != nil {
		//log.Errorf("<RPC Init> fail to register in address : %s\n", address)
		createLog(n.nodePtr.node.address, "network.Init", "rpc.Register", "Error", err.Error())
		return err
	}
	n.lis, err = net.Listen("tcp", address)
	if err != nil {
		//log.Errorf("<RPC Init> fail to listen in address : %s\n", address)
		createLog(n.nodePtr.node.address, "network.Init", "rpc.Listen", "Error", err.Error())
		return err
	}
	//log.Infof("<RPC Init> service start success in %s\n", address)
	createLog(n.nodePtr.node.address, "network.Init", "default", "Info", "")

	go WrappedAccept(n.serv, n.lis, n.nodePtr.node)
	return nil
}

func (n *network) ShutDown() error {
	n.nodePtr.node.quitSignal <- true
	err := n.lis.Close()
	if err != nil {
		//log.Errorf("<ShutDown> fail to close listener in address %s\n", n.nodePtr.node.address)
		createLog(n.nodePtr.node.address, "network.ShutDown", "Listener.Close", "Error", err.Error())
		return err
	}
	//log.Infof("<ShutDown> succeed in closing network in address %s\n", n.nodePtr.node.address)
	createLog(n.nodePtr.node.address, "network.ShutDown", "default", "Info", "")

	return nil
}

func GenerateClient(address string) (*rpc.Client, error) {
	if address == "" {
		//log.Warningf("<GetClient> IP address is nil\n")
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
			//log.Infof("<GetClient> Timeout to %s\n", address)
			createLog(address, "network.GenerateClient", "rpc.Dial", "Error", err.Error())
		}
	}
	return nil, err
}

func IfOnline(address string) bool {
	client, err := GenerateClient(address)
	if err != nil {
		//log.Infof("<IfOnline> Fail to connect to address: %s\n ", address)
		createLog(address, "network.IfOnline", "network.GenerateClient", "Error", err.Error())
		return false
	}
	if client != nil {
		defer client.Close()
	} else {
		return false
	}

	//}
	//log.Infof("<IfOnline> address: %s is on line\n ", address)
	//createLog(address, "network.IfOnline", "default", "Info", "")

	return true
}

func RemoteCall(addr string, serviceMethod string, args interface{}, reply interface{}) error {

	client, err := GenerateClient(addr)
	if err != nil {
		//log.Warningf("<RemoteCall> fail to generate client of address: %s\n", addr)
		createLog(addr, "network.RemoteCall", "network.GenerateClient", "Error", err.Error())
		return err
	}
	//createLog(addr, "network.RemoteCall", "network.GenerateClient", "Info", "after GenerateClient")
	if client != nil {
		defer client.Close()
	}
	//var str string
	//str = "1234"
	//createLog(addr, "network.RemoteCall", "client.Call.Hello", "Info", "before hello")
	//err = client.Call("WrapNode.Hello", str, str)

	//createLog(addr, "network.RemoteCall", "client.Call.Hello", "Info", "after hello")

	err = client.Call(serviceMethod, args, reply)
	//createLog(addr, "network.RemoteCall", "client.Call", "Info", "after call")
	if err != nil {
		//log.Warningf("<RemoteCall> client to %s fail to call method %s with error %s\n", addr, serviceMethod, err1)
		createLog(addr, "network.RemoteCall", "client.Call", "Error", err.Error())
		return err
	} else {
		//log.Warningf("<RemoteCall> client to %s succeed in calling method %s\n", addr, serviceMethod)
		//createLog(addr, "network.RemoteCall", "client.Call", "Info", "")

		return nil
	}
}

func WrappedAccept(server *rpc.Server, lis net.Listener, ptr *Node) {
	for {
		conn, err := lis.Accept()
		select {
		case <-ptr.quitSignal:
			return
		default:
			if err != nil {
				//log.Print("rpc.Serve: accept:", err.Error())
				createLog(ptr.address, "network.WrappedAccept", "listener.Accept", "Error", err.Error())
				return
			}
			go server.ServeConn(conn)
		}
	}
}
