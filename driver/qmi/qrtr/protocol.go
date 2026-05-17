package qrtr

const (
	portControl = 0xfffffffe
)

type packetType uint32

const (
	packetTypeData packetType = iota + 1
	packetTypeHello
	packetTypeBye
	packetTypeNewServer
	packetTypeDelServer
	packetTypeDelClient
	packetTypeResumeTx
	packetTypeExit
	packetTypePing
	packetTypeNewLookup
	packetTypeDelLookup
)

type sockAddr struct {
	Family uint16
	Node   uint32
	Port   uint32
}

type controlPacket struct {
	Command packetType
	Service service
}

type service struct {
	Service  uint32
	Instance uint32
	Node     uint32
	Port     uint32
}
