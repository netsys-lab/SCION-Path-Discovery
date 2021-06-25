package packets

// Some Metrics to start with
// Will be extended later
// NOTE: Add per path metrics here?
type PacketMetrics struct {
	ReadBytes      int64
	ReadPackets    int64
	WrittenBytes   int64
	WrittenPackets int64
}
