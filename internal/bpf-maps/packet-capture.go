package bpf_maps

type PktCounterValue struct {
	RxPackets uint64
	RxBytes   uint64
}

type PktCounterMapItem struct {
	Key   []uint8
	Value PktCounterValue
}
