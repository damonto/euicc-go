package apdu

type SmartCardChannel interface {
	Connect() error
	Disconnect() error
	OpenLogicalChannel(AID []byte) (byte, error)
	Transmit(command []byte) ([]byte, error)
	CloseLogicalChannel(channel byte) error
}
