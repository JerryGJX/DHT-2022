package bitTorrent



import "chord"

func NewNode(port int) dhtNode {
	ptr := new(chord.Node)
	ptr.Init(port)
	return ptr
}