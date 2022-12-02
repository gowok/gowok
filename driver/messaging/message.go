package messaging

type Message struct {
	Headers Table
	Tag     uint64
	Message []byte
}
