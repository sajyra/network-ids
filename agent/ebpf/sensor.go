//go:build linux && arm64

package ebpf

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"

	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
	"github.com/sanjay-rajjan/network-ids/pkg/types"
)

const (
	threatSYNFlood = 1
	threatPortScan = 2
	threatSSHBrute = 3
)

type Sensor struct {
	objs   sensorObjects
	link   link.Link
	reader *ringbuf.Reader
}

func NewSensor(ifaceName string) (*Sensor, error) {
	var objs sensorObjects
	if err := loadSensorObjects(&objs, nil); err != nil {
		return nil, fmt.Errorf("loading eBPF objects: %w", err)
	}

	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		objs.Close()
		return nil, fmt.Errorf("finding interface %s: %w", ifaceName, err)
	}

	l, err := link.AttachXDP(link.XDPOptions{
		Program:   objs.DetectThreats,
		Interface: iface.Index,
	})
	if err != nil {
		objs.Close()
		return nil, fmt.Errorf("attaching XDP: %w", err)
	}

	reader, err := ringbuf.NewReader(objs.ThreatEvents)
	if err != nil {
		l.Close()
		objs.Close()
		return nil, fmt.Errorf("creating ring buffer reader: %w", err)
	}

	log.Printf("[Sensor] Attached to interface %s", ifaceName)

	return &Sensor{
		objs:   objs,
		link:   l,
		reader: reader,
	}, nil

}

func (s *Sensor) ReadEvents(events chan<- types.ThreatEvent) {
	go func() {
		for {
			record, err := s.reader.Read()
			if err != nil {
				if err == ringbuf.ErrClosed {
					return
				}
				log.Printf("[Sensor] Error reading ring buffer: %v", err)
				continue
			}

			if len(record.RawSample) < 16 {
				continue
			}

			srcIP := binary.LittleEndian.Uint32(record.RawSample[0:4])
			threatType := binary.LittleEndian.Uint32(record.RawSample[4:8])
			timestamp := binary.LittleEndian.Uint64(record.RawSample[8:16])

			ip := make(net.IP, 4)
			binary.LittleEndian.PutUint32(ip, srcIP)

			var ttype types.ThreatType
			switch threatType {
			case threatSYNFlood:
				ttype = types.SYNFlood
			case threatPortScan:
				ttype = types.PortScan
			case threatSSHBrute:
				ttype = types.SSHBruteForce
			default:
				continue
			}

			events <- types.ThreatEvent{
				SourceIP:  ip.String(),
				Type:      ttype,
				Timestamp: int64(timestamp),
				NodeID:    "",
			}

		}
	}()
}
