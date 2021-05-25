package bpf_maps

type PktCounterKey struct {
	SourceAddr string
	DestAddr   string
	Family     uint32
}

type PktCounterValue struct {
	RxPackets uint64
	RxBytes   uint64
}

type PktCounterMapItem struct {
	Key   PktCounterKey
	Value PktCounterValue
}
