package bpf_maps

type PktCounterKey struct {
	SourceAddr []byte
	DestAddr   []byte
	Family     []byte
}

type PktCounterValue struct {
	RxPackets []byte
	RxBytes   []byte
}

type PktCounterMap struct {
	Key   PktCounterKey
	Value PktCounterValue
}
