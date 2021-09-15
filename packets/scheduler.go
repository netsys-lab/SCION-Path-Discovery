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
	SetConnections([]UDPConn)
}

func NewWeighedScheduler(local snet.UDPAddr) *SampleFirstPathScheduler {
	weighedScheduler := SampleFirstPathScheduler{
		local: local,
	}

	return &weighedScheduler
}

// Implements a PacketScheduler that calculates weights out of
// PathQualities and sends packets depending on the weight of
// each alternative
type SampleFirstPathScheduler struct {
	connections []UDPConn
	local       snet.UDPAddr
}

/*func (ws *WeighedScheduler) Accept() (*peers.PathlevelPeer, error) {
	conn := ws.transportConstructor()
	err := conn.Listen(ws.local)
	if err != nil {
		return nil, err
	}

	peer, err := conn.Accept()
	if err != nil {
		return nil, err
	}

	return peer, nil
}*/

func (ws *SampleFirstPathScheduler) SetConnections(conns []UDPConn) {
	ws.connections = conns
}

/*func (ws *WeighedScheduler) SetPathlevelPeers(peers []peers.PathlevelPeer) error {
	// TODO: Keep same connections open
	for _, conn := range ws.connections {
		err := conn.Close()
		if err != nil {
			return err
		}
	}

	ws.connections = make([]TransportConn, 0)

	for _, peer := range peers {
		transportConn := ws.transportConstructor()
		transportConn.Connect(peer.PeerAddr, &peer.PathQuality.Path)
		ws.connections = append(ws.connections, transportConn)
	}

	return nil
}*/

func (ws *SampleFirstPathScheduler) Write(data []byte) (int, error) {
	if len(ws.connections) < 2 {
		return 0, errors.New("No connection available to write")
	}
	return ws.connections[1].Write(data)
}
func (ws *SampleFirstPathScheduler) Read(data []byte) (int, error) {
	if len(ws.connections) < 1 {
		return 0, errors.New("No connection available to read")
	}
	// We assume that the first conn here is always the one that was initialized by listen()
	// Other cons could be added due to handshakes (QUIC specific)
	return ws.connections[0].Read(data)
}
func (ws *SampleFirstPathScheduler) WriteStream(data []byte) (int, error) {
	if len(ws.connections) < 2 {
		return 0, errors.New("No connection available to writeStrean")
	}
	return ws.connections[1].WriteStream(data)
}
func (ws *SampleFirstPathScheduler) ReadStream(data []byte) (int, error) {
	if len(ws.connections) < 1 {
		return 0, errors.New("No connection available to readStream")
	}
	// We assume that the first conn here is always the one that was initialized by listen()
	// Other cons could be added due to handshakes (QUIC specific)
	return ws.connections[0].ReadStream(data)
}
