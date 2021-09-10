package packets

import (
	"errors"

	"github.com/scionproto/scion/go/lib/snet"
)

// PacketSchedulers have a list of "selected" paths which comes
// from the path selection module. These paths are considered
// "optimal" regarding the path selection properties that
// the application provides
// The scheduler decides for each packet over which of the optimal
// paths it will be sent
type PacketScheduler interface {
	Write([]byte) (int, error)
	Read([]byte) (int, error)
	WriteStream([]byte) (int, error)
	ReadStream([]byte) (int, error)
	SetListenConnections([]UDPConn)
	SetDialConnections([]UDPConn)
}

func NewSampleFirstPath(local snet.UDPAddr) *SampleFirstPathScheduler {
	weighedScheduler := SampleFirstPathScheduler{
		local: local,
	}

	return &weighedScheduler
}

// Implements a PacketScheduler that calculates weights out of
// PathQualities and sends packets depending on the weight of
// each alternative
type SampleFirstPathScheduler struct {
	listenConnections []UDPConn
	local             snet.UDPAddr
	dialConnections   []UDPConn
}

func (ws *SampleFirstPathScheduler) SetDialConnections(conns []UDPConn) {
	ws.dialConnections = conns
}

// TODO: Filter identical connections?
func (ws *SampleFirstPathScheduler) SetListenConnections(conns []UDPConn) {
	ws.listenConnections = conns
}

func (ws *SampleFirstPathScheduler) Write(data []byte) (int, error) {
	if len(ws.dialConnections) < 1 {
		return 0, errors.New("No connection available to write")
	}
	return ws.dialConnections[0].Write(data)
}
func (ws *SampleFirstPathScheduler) Read(data []byte) (int, error) {
	if len(ws.listenConnections) < 1 {
		return 0, errors.New("No connection available to read")
	}
	// We assume that the first conn here is always the one that was initialized by listen()
	// Other cons could be added due to handshakes (QUIC specific)
	return ws.listenConnections[0].Read(data)
}
func (ws *SampleFirstPathScheduler) WriteStream(data []byte) (int, error) {
	if len(ws.dialConnections) < 1 {
		return 0, errors.New("No connection available to writeStrean")
	}
	return ws.dialConnections[0].WriteStream(data)
}
func (ws *SampleFirstPathScheduler) ReadStream(data []byte) (int, error) {
	if len(ws.listenConnections) < 1 {
		return 0, errors.New("No connection available to readStream")
	}
	// We assume that the first conn here is always the one that was initialized by listen()
	// Other cons could be added due to handshakes (QUIC specific)
	return ws.listenConnections[0].ReadStream(data)
}
