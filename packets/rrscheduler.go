package packets

import (
	"errors"

	"github.com/scionproto/scion/go/lib/snet"
)

func NewRoundRobinScheduler(local snet.UDPAddr) *RoundRobinScheduler {
	weighedScheduler := RoundRobinScheduler{
		local: local,
	}

	return &weighedScheduler
}

// Implements a PacketScheduler that calculates weights out of
// PathQualities and sends packets depending on the weight of
// each alternative
type RoundRobinScheduler struct {
	listenConnections []UDPConn
	local             snet.UDPAddr
	dialConnections   []UDPConn
}

func (ws *RoundRobinScheduler) SetDialConnections(conns []UDPConn) {
	ws.dialConnections = conns
}

// TODO: Filter identical connections?
func (ws *RoundRobinScheduler) SetListenConnections(conns []UDPConn) {
	ws.listenConnections = conns
}

func (ws *RoundRobinScheduler) Write(data []byte) (int, error) {
	if len(ws.dialConnections) < 1 {
		return 0, errors.New("No connection available to write")
	}
	return ws.dialConnections[0].Write(data)
}
func (ws *RoundRobinScheduler) Read(data []byte) (int, error) {
	if len(ws.listenConnections) < 1 {
		return 0, errors.New("No connection available to read")
	}
	// We assume that the first conn here is always the one that was initialized by listen()
	// Other cons could be added due to handshakes (QUIC specific)
	return ws.listenConnections[0].Read(data)
}
func (ws *RoundRobinScheduler) WriteStream(data []byte) (int, error) {
	if len(ws.dialConnections) < 1 {
		return 0, errors.New("No connection available to writeStrean")
	}
	return ws.dialConnections[0].WriteStream(data)
}
func (ws *RoundRobinScheduler) ReadStream(data []byte) (int, error) {
	if len(ws.listenConnections) < 1 {
		return 0, errors.New("No connection available to readStream")
	}
	// We assume that the first conn here is always the one that was initialized by listen()
	// Other cons could be added due to handshakes (QUIC specific)
	return ws.listenConnections[0].ReadStream(data)
}
