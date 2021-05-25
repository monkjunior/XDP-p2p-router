package packet_capture

type PacketInfo struct {
	SourceAddr []byte
	DestAddr   []byte
	Family     []byte
}

type PacketCounter struct {
	RxPackets []byte
	RxBytes   []byte
}
