#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/in.h>
#include <linux/tcp.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>

#define MAX_ENTRIES 1024
#define SYN_THRESHOLD  100
#define PORT_THRESHOLD  200
#define SSH_THRESHOLD   10
#define SSH_PORT  22

struct threat_event {
    __u32 src_ip;
    __u32 threat_type;
    __u64 timestamp;
};

#define THREAT_SYN_FLOOD  1
#define THREAT_PORT_SCAN  2
#define THREAT_SSH_BRUTE  3

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, MAX_ENTRIES);
    __type(key, __u32);
    __type(value, __u32);
} blocked_ips SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, MAX_ENTRIES);
    __type(key, __u32);
    __type(value, __u32);
} syn_count SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, MAX_ENTRIES);
    __type(key, __u32);
    __type(value, __u32);
} port_count SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, MAX_ENTRIES);
    __type(key, __u32);
    __type(value, __u32);
} ssh_count SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 1<<16);
} threat_events SEC(".maps");

SEC("xdp")
int detect_threats(struct xdp_md *ctx) {
    void* data = (void*) (long) ctx->data;
    void* data_end = (void*)(long) ctx->data_end;

    struct ethhdr *eth = data;
    //packet too short for full ethernet header
    if ((void*)(eth + 1) > data_end) {
        return XDP_PASS;
    }
    // check if IPv4 packet
    if (eth->h_proto != bpf_htons(ETH_P_IP)) {
        return XDP_PASS;
    }
    struct iphdr *ip = (void*)(eth + 1);
    if ((void*)(ip + 1) > data_end) {
        return XDP_PASS;
    }

    if (ip->protocol != IPPROTO_TCP) {
        return XDP_PASS;
    }

    __u32 src_ip = ip->saddr;

    struct tcphdr *tcp = (void*)(ip + 1);
    if ((void*)(tcp + 1) > data_end) {
        return XDP_PASS;
    }

    __u32 *blocked = bpf_map_lookup_elem(&blocked_ips, &src_ip);
    if (blocked) {
        return XDP_DROP;
    }

    __u16 dst_port = bpf_ntohs(tcp->dest);
    __u32 one = 1;
    __u32 *count;

    if (tcp->syn && !tcp->ack) {
        count = bpf_map_lookup_elem(&syn_count, &src_ip);
        if (count) {
            __sync_fetch_and_add(count, 1);
            if (*count > SYN_THRESHOLD) {
                struct threat_event *e = bpf_ringbuf_reserve(&threat_events, sizeof(*e), 0);
                if (e) {
                    e->src_ip = src_ip;
                    e->threat_type = THREAT_SYN_FLOOD;
                    e->timestamp = bpf_ktime_get_ns();
                    bpf_ringbuf_submit(e, 0);
                }
            }
        } else {
            bpf_map_update_elem(&syn_count, &src_ip, &one, BPF_ANY);
        }
    }
    count = bpf_map_lookup_elem(&port_count, &src_ip);
    if (count) {
        __sync_fetch_and_add(count, 1);
        if (*count > PORT_THRESHOLD) {
            struct threat_event *e = bpf_ringbuf_reserve(&threat_events, sizeof(*e), 0);
            if (e) {
                e->src_ip      = src_ip;
                e->threat_type = THREAT_PORT_SCAN;
                e->timestamp   = bpf_ktime_get_ns();
                bpf_ringbuf_submit(e, 0);
            }
        }
    } else {
        bpf_map_update_elem(&port_count, &src_ip, &one, BPF_ANY);
    }

    if (dst_port == SSH_PORT && tcp->syn && !tcp->ack) {
        count = bpf_map_lookup_elem(&ssh_count, &src_ip);
        if (count) {
            __sync_fetch_and_add(count, 1);
            if (*count > SSH_THRESHOLD) {
                struct threat_event *e = bpf_ringbuf_reserve(&threat_events, sizeof(*e), 0);
                if (e) {
                    e->src_ip      = src_ip;
                    e->threat_type = THREAT_SSH_BRUTE;
                    e->timestamp   = bpf_ktime_get_ns();
                    bpf_ringbuf_submit(e, 0);
                }
            }
        } else {
            bpf_map_update_elem(&ssh_count, &src_ip, &one, BPF_ANY);
        }
    }

    return XDP_PASS;

}

char LICENSE[] SEC("license") = "GPL";






