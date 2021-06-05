package bpf_maps

type PktCounterValue struct {
	RxPackets uint64
	RxBytes   uint64
}

type PktCounterMapItem struct {
	Key   uint32
	Value PktCounterValue
}
