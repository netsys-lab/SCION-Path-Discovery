package socket

import (
	"github.com/netsys-lab/scion-path-discovery/packets"
	"github.com/scionproto/scion/go/lib/snet"
)

var _ UnderlaySocket = (*SCIONSocket)(nil)

// TODO: extend this further. It may be useful to use more than
// one native UDP socket due to performance limitations
//type Socket interface {
//	net.Conn
//}

type SCIONSocket struct {
	ctrlConn             packets.UDPConn
	local                string
	localAddr            *snet.UDPAddr
	transportConstructor packets.TransportConstructor
	conns                []packets.UDPConn
}

func NewSCIONSocket(local string, transportConstructor packets.TransportConstructor) *SCIONSocket {
	s := SCIONSocket{
		local:                local,
		transportConstructor: transportConstructor,
	}

	return &s
}

func (s *SCIONSocket) Listen() error {
	lAddr, err := snet.ParseUDPAddr(s.local)
	if err != nil {
		return err
	}

	s.localAddr = lAddr
	s.ctrlConn = s.transportConstructor()
	return s.ctrlConn.Listen(*s.localAddr)
}

func (s *SCIONSocket) WaitForDialIn() (packets.UDPConn, *snet.UDPAddr, error) {
	// TODO: Close
	bytes := make([]byte, packets.PACKET_SIZE)
	_, err := s.ctrlConn.Read(bytes)
	if err != nil {
		return nil, nil, err
	}

	// TODO: Handle Packets appropriate
	return nil, nil, err
}

func (s *SCIONSocket) AcceptAll() (*snet.UDPAddr, []packets.UDPConn, error) {
	// TODO: Close
	bytes := make([]byte, packets.PACKET_SIZE)
	_, err := s.ctrlConn.Read(bytes)
	if err != nil {
		return nil, nil, err
	}

	// TODO: Handle Packets appropriate
	return nil, make([]packets.UDPConn, 0), nil
}

func (s *SCIONSocket) Dial(remote snet.UDPAddr, path snet.Path) (packets.UDPConn, error) {
	// TODO: Handle Packets appropriate
	// TODO: Append connection to close them later
	return nil, nil
}

func (s *SCIONSocket) DialAll(remote snet.UDPAddr, path []snet.Path) ([]packets.UDPConn, error) {
	// TODO: Handle Packets appropriate
	// TODO: Append connections to close them later
	return make([]packets.UDPConn, 0), nil
}

func (s *SCIONSocket) GetConnections() []packets.UDPConn {
	return s.conns
}

func (s *SCIONSocket) CloseAll() []error {
	errors := make([]error, 0)
	for _, con := range s.conns {
		err := con.Close()
		if err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}
