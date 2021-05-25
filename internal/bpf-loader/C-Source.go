package bpf_loader

const (
	C_SOURCE_CODE = `
/* SPDX-License-Identifier: GPL-2.0 */
#include <linux/bpf.h>
#include <uapi/linux/if_ether.h>
#include <linux/in.h>
#include <linux/ip.h>

/* packetInfo struct is used as key for our eBPF map */
struct packet_info {
	uint32_t s_v4_addr;
	uint32_t d_v4_addr;
	__u8 family;
};

/* packetCounter struct is used as value for our eBPF map */
struct packet_counter {
    __u64 rx_packets;
    __u64 rx_bytes;
};

BPF_TABLE("hash", struct packet_info, struct packet_counter, counters, 1024);

int packet_counter(struct xdp_md *ctx){

	void* data_end = (void*)(long)ctx->data_end;
	void* data     = (void*)(long)ctx->data;

	struct ethhdr *eth = data;
    
	struct packet_info key = {};
	struct packet_counter default_value = {};
	struct packet_counter *value;

	int ipSize = 0;

	struct iphdr *ip;

	ipSize = sizeof(*eth);
	ip = data + ipSize;
	ipSize += sizeof(struct iphdr);

	if (data + ipSize > data_end) {
		return XDP_DROP;
	}

	key.family = ip->protocol;
	key.s_v4_addr = ip->saddr;
	key.d_v4_addr = ip->daddr;

	if (ip->daddr==LOCAL_ADDR) {
		value = counters.lookup_or_try_init(&key, &default_value);
		if (value) {
			__u64 bytes = data_end - data; /* Calculate packet length */
			__sync_fetch_and_add(&value->rx_packets, 1);
			__sync_fetch_and_add(&value->rx_bytes, bytes);
		}
	}

	return XDP_PASS;
}
`
)
