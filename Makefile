.PHONY: generate build run stop clean

generate:
	cd agent/ebpf && \
	GOPACKAGE=ebpf bpf2go -target arm64 -output-dir . sensor \
		bpf/sensor.c -- \
		-O2 -g \
		-D__TARGET_ARCH_arm64 \
		-I/usr/include/aarch64-linux-gnu

build-coordinator:
	go build -o bin/coordinator ./coordinator/

build-agent:
	cd /tmp && go build -o agent \
		github.com/sanjay-rajjan/network-ids/agent

run:
	cd docker && docker compose up --build

stop:
	cd docker && docker compose down

clean:
	rm -f bin/coordinator
	rm -f /tmp/agent