package bpf_loader

const (
	CSourceCode = `
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

BPF_TABLE("hash", uint32_t, struct packet_counter, pkt_counter, 1024);
BPF_TABLE("hash", uint32_t, int, ip_whitelist, 1024);

int packet_counter(struct xdp_md *ctx){

	void* data_end = (void*)(long)ctx->data_end;
	void* data     = (void*)(long)ctx->data;

	struct ethhdr *eth = data;
    
	uint32_t pkt_cap_key;
	struct packet_counter pkt_cap_dft_value = {};
	struct packet_counter *pkt_cap_value;
	
	uint32_t ip_whitelist_key;
	int ip_whitelist_dft_value = XDP_PASS;
	int *ip_whitelist_value;

	int ipSize = 0;

	struct iphdr *ip;

	ipSize = sizeof(*eth);
	ip = data + ipSize;
	ipSize += sizeof(struct iphdr);

	if (data + ipSize > data_end) {
		return XDP_DROP;
	}

	pkt_cap_key= ip->saddr;

	//if (ip->daddr==LOCAL_ADDR) {
	//	pkt_cap_value = pkt_counter.lookup_or_try_init(&pkt_cap_key, &pkt_cap_dft_value);
	//	if (pkt_cap_value) {
	//		__u64 bytes = data_end - data; /* Calculate packet length */
	//		__sync_fetch_and_add(&pkt_cap_value->rx_packets, 1);
	//		__sync_fetch_and_add(&pkt_cap_value->rx_bytes, bytes);
	//	}
	//	
	//	ip_whitelist_key = ip->saddr;
	//	ip_whitelist_value = ip_whitelist.lookup_or_init(&ip_whitelist_key, &ip_whitelist_dft_value);
	//	if (ip_whitelist_value) {
	//		return *ip_whitelist_value;
	//	}
	//}

	pkt_cap_value = pkt_counter.lookup_or_try_init(&pkt_cap_key, &pkt_cap_dft_value);
	if (pkt_cap_value) {
		__u64 bytes = data_end - data; /* Calculate packet length */
		__sync_fetch_and_add(&pkt_cap_value->rx_packets, 1);
		__sync_fetch_and_add(&pkt_cap_value->rx_bytes, bytes);
	}
	
	ip_whitelist_key = ip->saddr;
	ip_whitelist_value = ip_whitelist.lookup_or_init(&ip_whitelist_key, &ip_whitelist_dft_value);
	if (ip_whitelist_value) {
		return *ip_whitelist_value;
	}

	return XDP_PASS;
}
`
)
