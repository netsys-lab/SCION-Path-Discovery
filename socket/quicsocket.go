package socket

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/netsys-lab/scion-path-discovery/packets"
	"github.com/scionproto/scion/go/lib/snet"
)

var _ packets.UDPConn = (*packets.QUICReliableConn)(nil)

var _ UnderlaySocket = (*QUICSocket)(nil)

// TODO: extend this further. It may be useful to use more than
// one native UDP socket due to performance limitations
//type Socket interface {
//	net.Conn
//}

type QUICSocket struct {
	listenConns          []*packets.QUICReliableConn
	local                string
	localAddr            *snet.UDPAddr
	transportConstructor packets.TransportConstructor
	dialConns            []*packets.QUICReliableConn
}

func NewQUICSocket(local string) *QUICSocket {
	s := QUICSocket{
		local:       local,
		listenConns: make([]*packets.QUICReliableConn, 0),
		dialConns:   make([]*packets.QUICReliableConn, 0),
	}

	return &s
}

func (s *QUICSocket) Listen() error {
	lAddr, err := snet.ParseUDPAddr(s.local)
	if err != nil {
		return err
	}

	s.localAddr = lAddr
	conn := packets.QUICConnConstructor()
	s.listenConns = append(s.listenConns, conn)
	return conn.Listen(*s.localAddr)
}

func (s *QUICSocket) WaitForDialIn() (*snet.UDPAddr, error) {
	bts := make([]byte, packets.PACKET_SIZE)
	stream, err := s.listenConns[0].AcceptStream()
	if err != nil {
		return nil, err
	}

	_, err = stream.Read(bts)
	fmt.Println("Read something")
	if err != nil {
		return nil, err
	}
	p := DialPacket{}
	network := bytes.NewBuffer(bts) // Stand-in for a network connection
	dec := gob.NewDecoder(network)
	err = dec.Decode(&p)
	if err != nil {
		return nil, err
	}

	addr := p.Addr

	return &addr, nil
}

func (s *QUICSocket) Dial(remote snet.UDPAddr, path snet.Path, options DialOptions) (packets.UDPConn, error) {
	// appnet.SetPath(&remote, path)
	// fmt.Printf("Dialing to %s via %s\n", remote.String(), remote.Path)
	conn := packets.QUICConnConstructor()
	conn.SetLocal(*s.localAddr)
	err := conn.Dial(remote, &path)
	if err != nil {
		return nil, err
	}

	if options.SendAddrPacket {
		var network bytes.Buffer
		enc := gob.NewEncoder(&network) // Will write to network.
		p := DialPacket{
			Addr: remote,
		}

		err := enc.Encode(p)
		conn.Write(network.Bytes())
		if err != nil {
			return nil, err
		}
	}

	s.dialConns = append(s.dialConns, conn)

	return conn, nil
}

func (s *QUICSocket) DialAll(remote snet.UDPAddr, path []snet.Path, options DialOptions) ([]packets.UDPConn, error) {
	conns := make([]packets.UDPConn, 0)
	for _, v := range path {
		conn, err := s.Dial(remote, v, options)
		if err != nil {
			return nil, err
		}
		conns = append(conns, conn)
	}
	fmt.Println("Dial all#1")

	return conns, nil
}

func (s *QUICSocket) GetListenConnections() []packets.UDPConn {
	conns := make([]packets.UDPConn, 0)
	for _, v := range s.listenConns {
		conns = append(conns, v)
	}
	return conns
}

func (s *QUICSocket) GetDialConnections() []packets.UDPConn {
	conns := make([]packets.UDPConn, 0)
	for _, v := range s.dialConns {
		conns = append(conns, v)
	}
	return conns
}

func (s *QUICSocket) CloseAll() []error {
	errors := make([]error, 0)
	for _, con := range s.dialConns {
		err := con.Close()
		if err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}
