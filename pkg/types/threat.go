package types

type ThreatType string

const (
	SYNFlood ThreatType = "SYN_FLOOD"
	PortScan ThreatType = "PORT_SCAN"
	SSHBruteForce ThreatType = "SSH_BRUTE_FORCE"
)

type ThreatEvent struct {
	SourceIP string
	Type ThreatType
	Timestamp int64
	NodeID string
}