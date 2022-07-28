package bitTorrent

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"time"

	bencode "github.com/jackpal/bencode-go"
)

type Client struct {
	addr string
	node dhtNode
}

func (ptr *Client) Init(port int) {
	ptr.node = NewNode(port)
	ptr.node.Run()
	yellow.Printf("Client start running at %s:%d\n", GetLocalIP(), port)
}

func (ptr *Client) Create() {
	ptr.node.Create()
	green.Println("[success] Create new network.")
}

func (ptr *Client) Join(address string) {
	ok := ptr.node.Join(address)
	if ok {
		green.Printf("[success] Joined %s.\n", address)
	} else {
		red.Printf("[fail] Join %s failed.\n", address)
	}
}

func (ptr *Client) Quit() {
	ptr.node.Quit()
	green.Println("[success] Quit from current network.")
}


func (c *Client) DownLoadByMagnet(magnet string, savePath string) {
	ok, torrentStr := c.node.Get(magnet)
	if ok {
		reader := bytes.NewBufferString(torrentStr)
		torrent := bencodeTorrent{}
		err := bencode.Unmarshal(reader, &torrent)
		if err != nil {
			red.Printf("Unmarshal torrent error: %v\n", err)
		} else {
			torrent.Save("tmp.torrent")
			green.Println("Save torrent file to tmp.torrent.")
		}
	} else {
		red.Println("Cannot find torrent of given magnet URL.")
	}
	c.DownLoadByTorrent("tmp.torrent", savePath)
}

func (ptr *Client) DownLoadByTorrent(torrentPath string, savePath string) {
	fileIO, err := os.Create(savePath)
	if err != nil {
		red.Printf("Create file %s error: %v.\n", savePath, err)
		return
	}
	torrent, err := Open(torrentPath)
	if err != nil {
		red.Printf("Failed to open torrent in path %s, error: %v.\n", torrentPath, err.Error())
		return
	}
	key := torrent.makeKeyPackage()
	yellow.Printf("File name: %s, size: %v bytes, piece num: %v.\n", torrent.Info.Name, torrent.Info.Length, key.size)
	fileName := torrent.Info.Name

	ok, data := ptr.DownLoad(&key)
	if ok {
		fileIO.Write(data)
		green.Printf("Download %s to %s finished.", fileName, savePath)
	} else {
		red.Println("Download failed.")
	}
}

func (this *Client) Upload(filePath,torrentPath string) {
	yellow.Printf("Start uploading file \n")
	dataPackage, length := makeDataPackage(filePath)
	green.Println("Data packaged, piece number: ", dataPackage.size)

	var pieces string
	for i := 0; i < dataPackage.size; i++ {
		piece, _ := PiecesHash(dataPackage.data[i], i)
		pieces += fmt.Sprintf("%x", piece)
	}
	_, fileName := path.Split(filePath)
	torrent := bencodeTorrent{
		Announce: "",
		Info: bencodeInfo{
			Length:      length,
			Pieces:      pieces,
			PieceLength: PieceSize,
			Name:        fileName,
		},
	}
	err := torrent.Save(torrentPath)
	if err != nil {
		red.Printf("[fail] Save torrent file error: %v.\n", err)
		return
	}

	key := torrent.makeKeyPackage()
	yellow.Printf("File name: %s, size: %v bytes, piece num: %v.\n", torrent.Info.Name, torrent.Info.Length, key.size)

	var buf bytes.Buffer
	err = bencode.Marshal(&buf, torrent)
	if err != nil {
		red.Printf("[fail] Torrent marshal error: %v.\n", err)
	}
	magnet := MakeMagnet(fmt.Sprintf("%x", key.infoHash))
	torrentStr := buf.String()

	ok := this.UploadPackage(&key, &dataPackage)
	if ok {
		yellow.Printf("Upload finished, create torrent file: %s\n", torrentPath)
		this.node.Put(magnet, torrentStr)
		yellow.Printf("Magnet URL: %s\n", magnet)
	}
}

type Worker struct {
	index   int
	retry   int
	success bool
	result  DataPiece
}

func (this *Client) uploadPieceWork(keyPackage *KeyPackage, dataPackage *DataPackage, index int, retry int, workQueue *chan Worker) {
	key := keyPackage.getKey(index)
	ok := this.node.Put(key, string(dataPackage.data[index]))
	if ok {
		*workQueue <- Worker{index: index, success: true, retry: retry}
	} else {
		*workQueue <- Worker{index: index, success: false, retry: retry}
	}
}

func (c *Client) UploadPackage(key *KeyPackage, data *DataPackage) bool {
	workQueue := make(chan Worker, WorkQueueBuffer)
	for i := 0; i < key.size; i++ {
		go c.uploadPieceWork(key, data, i, 0, &workQueue)
	}
	donePieces := 0
	for donePieces < key.size {
		select {
		case work := <-workQueue:
			if work.success {
				donePieces++
				green.Printf("Piece #%d uploaded.\n", work.index+1)
			} else {
				red.Printf("Piece #%d upload failed %d/%d.\n", work.index+1, work.retry, RetryTime)
				if work.retry < RetryTime {
					go c.uploadPieceWork(key, data, work.index, work.retry+1, &workQueue)
				} else {
					red.Println("Upload failed.")
					return false
				}
			}
		case <-time.After(UploadTimeout * time.Duration(key.size)):
			red.Println("Upload timeout.")
			return false
		}
	}
	green.Println("Upload finished.")
	return true
}

func (this *Client) downloadPieceWork(keyPackage *KeyPackage, index int, retry int, workQueue *chan Worker) {
	key := keyPackage.getKey(index)
	ok, piece := this.node.Get(key)
	if ok {
		*workQueue <- Worker{index: index, success: true, result: []byte(piece), retry: retry}
	} else {
		*workQueue <- Worker{index: index, success: false, retry: retry}
	}
}

func (this *Client) DownLoad(keyPackage *KeyPackage) (bool, []byte) {
	ret := make([]byte, keyPackage.length)
	workQueue := make(chan Worker, WorkQueueBuffer)
	for i := 0; i < keyPackage.size; i++ {
		go this.downloadPieceWork(keyPackage, i, 0, &workQueue)
	}

	var cnt int

	checker := make([]KeyPiece, keyPackage.size)

	for cnt != keyPackage.size {
		select {
		case worker := <-workQueue:
			{
				if worker.success {
					cnt++
					bound := (worker.index + 1) * PieceSize
					if bound > keyPackage.length {
						bound = keyPackage.length
					}
					copy(ret[worker.index*PieceSize:bound], worker.result)
					pieceHash, _ := PiecesHash(worker.result, worker.index)
					checker[worker.index] = pieceHash
					if pieceHash != keyPackage.key[worker.index] {
						red.Printf("Check integrity of Piece #%d failed.\n", worker.index+1)
						return false, []byte{}
					}
					yellow.Printf("Piece #%v Download Finish. (%.2f", worker.index+1, float64(cnt*100)/float64(keyPackage.size))
					yellow.Println("%)")
				} else {
					red.Printf("Piece #%v Download Failed. Retry Times: %d (%.2f", worker.index+1, worker.retry, float64(cnt*100)/float64(keyPackage.size))
					red.Println("%)")
					if worker.retry < RetryTime {
						go this.downloadPieceWork(keyPackage, worker.index, worker.retry+1, &workQueue)
					} else {
						red.Println("Download Failed: Retry Too Much, Killed")
						return false, []byte{}
					}
				}
			}
		case <-time.After(time.Duration(int64(DownloadTimeout) * int64(keyPackage.size))):
			{
				red.Println("Download Failed: TimeOut!")
				return false, []byte{}
			}
		}
	}

	green.Println("Download Data Success")

	return true, ret
}
