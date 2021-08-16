package packets

import (
	"github.com/netsys-lab/scion-path-discovery/peers"
	"github.com/scionproto/scion/go/lib/snet"
)

// PacketSchedulers have a list of "selected" paths which comes
// from the path selection module. These paths are considered
// "optimal" regarding the path selection properties that
// the application provides
// The scheduler decides for each packet over which of the optimal
// paths it will be sent
type PacketScheduler interface {
	SetPathlevelPeers(peers []peers.PathlevelPeer)
	SetPathQualities([]snet.Path) error
	Write([]byte) (int, error)
	Read([]byte) (int, error)
	WriteStream([]byte) (int, error)
	ReadStream([]byte) (int, error)
	SetConnections([]TransportConn)
	AddConnection(TransportConn)
}

func NewWeighedScheduler(local snet.UDPAddr) *WeighedScheduler {
	weighedScheduler := WeighedScheduler{
		local: local,
	}

	return &weighedScheduler
}

// Implements a PacketScheduler that calculates weights out of
// PathQualities and sends packets depending on the weight of
// each alternative
type WeighedScheduler struct {
	connections []TransportConn
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

func (ws *WeighedScheduler) SetConnections(conns []TransportConn) {
	ws.connections = conns
}

// TODO: Filter identical connections?
func (ws *WeighedScheduler) AddConnection(conn TransportConn) {
	ws.connections = append(ws.connections, conn)
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

func (ws *WeighedScheduler) Write([]byte) (int, error) {
	return 0, nil
}
func (ws *WeighedScheduler) Read([]byte) (int, error) {
	return 0, nil
}
func (ws *WeighedScheduler) WriteStream([]byte) (int, error) {
	return 0, nil
}
func (ws *WeighedScheduler) ReadStream([]byte) (int, error) {
	return 0, nil
}
