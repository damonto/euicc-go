package apdu

type SmartCardChannel interface {
	Connect() error
	Disconnect() error
	OpenLogicalChannel(aid []byte) (byte, error)
	Transmit([]byte) ([]byte, error)
	CloseLogicalChannel(byte) error
}
