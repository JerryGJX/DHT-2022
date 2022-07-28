# DHT-2022
Distributed Hash Table - for PPCA 2022


# Report：

## Chord & Kademlia finished
采用的第三方库：
```
github.com/sirupsen/logrus
github.com/sasha-s/go-deadlock
```

## Application remain works to do
目前基本完成,有待调试
采用的第三方库：
```
github.com/sirupsen/logrus
github.com/sasha-s/go-deadlock
github.com/jackpal/bencode-go
github.com/fatih/color
```

期望实现torrent文件上传下载


## Reference

### for Chord
[Chord: A Scalable Peer-to-peer Lookup Protocol for Internet Applications](./ref/paper-ton.pdf)

[聊聊分布式散列表（DHT）的原理——以 Kademlia（Kad） 和 Chord 为例](https://program-think.blogspot.com/2017/09/Introduction-DHT-Kademlia-Chord.html)

### for Kademlia
[KADEMLIA算法学习](https://shuwoom.com/?p=813)

[The Kademlia Protocol Succinctly](https://www.syncfusion.com/succinctly-free-ebooks/kademlia-protocol-succinctly)


### for torrent
[Building a BitTorrent client from the ground up in Go](https://blog.jse.li/posts/torrent/#putting-it-all-together)