package socket

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/netsys-lab/scion-path-discovery/packets"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/lib/snet/path"
	log "github.com/sirupsen/logrus"
)

var _ packets.UDPConn = (*packets.QUICReliableConn)(nil)

var _ UnderlaySocket = (*QUICSocket)(nil)

type DialPacketQuic struct {
	Addr snet.UDPAddr
	// Path snet.Path
	NumPaths int
}

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
	acceptedConns        chan []*packets.QUICReliableConn
	options              *SockOptions
}

func NewQUICSocket(local string, opts *SockOptions) *QUICSocket {
	s := QUICSocket{
		local:         local,
		listenConns:   make([]*packets.QUICReliableConn, 0),
		dialConns:     make([]*packets.QUICReliableConn, 0),
		acceptedConns: make(chan []*packets.QUICReliableConn, 0),
		options:       opts,
	}

	gob.Register(path.Path{})

	return &s
}

func (s *QUICSocket) Listen() error {
	lAddr, err := snet.ParseUDPAddr(s.local)
	if err != nil {
		return err
	}

	s.localAddr = lAddr
	conn := &packets.QUICReliableConn{}
	s.listenConns = append(s.listenConns, conn)
	err = conn.Listen(*s.localAddr)

	return err
}

func (s *QUICSocket) WaitForIncomingConn() (packets.UDPConn, error) {
	if s.options == nil || !s.options.MultiportMode {
		log.Infof("Waiting for new connection")
		stream, err := s.listenConns[0].AcceptStream()
		if err != nil {
			log.Fatalf("QUIC Accept err %s", err.Error())
		}

		log.Debugf("Accepted new Stream on listen socket")

		bts := make([]byte, packets.PACKET_SIZE)
		n, err := stream.Read(bts)

		log.Debugf("Got %d bytes from new accepted stream", n)

		if s.listenConns[0].GetInternalConn() == nil {
			log.Debugf("Set stream to listen conn")
			s.listenConns[0].SetStream(stream)
			select {
			case s.listenConns[0].Ready <- true:
			default:
			}

			log.Debugf("Set connection ready")
			return s.listenConns[0], nil
		} else {
			newConn := &packets.QUICReliableConn{}
			newConn.SetLocal(*s.localAddr)
			newConn.SetRemote(s.listenConns[0].GetRemote())
			newConn.SetStream(stream)
			s.listenConns = append(s.listenConns, newConn)

			_, err = stream.Read(bts)
			if err != nil {
				return nil, err
			}
			return newConn, nil
		}
	} else {
		addr := s.localAddr.Copy()
		addr.Host.Port = s.localAddr.Host.Port + len(s.listenConns)
		conn := &packets.QUICReliableConn{}
		err := conn.Listen(*addr)
		if err != nil {
			return nil, err
		}

		stream, err := conn.AcceptStream()
		if err != nil {
			return nil, err
		}

		conn.SetStream(stream)
		s.listenConns = append(s.listenConns, conn)
		bts := make([]byte, packets.PACKET_SIZE)
		_, err = stream.Read(bts)
		if err != nil {
			return nil, err
		}
		return conn, nil
	}
}

func (s *QUICSocket) WaitForDialIn() (*snet.UDPAddr, error) {
	bts := make([]byte, packets.PACKET_SIZE)
	log.Debugf("Wait for Dial In")
	stream, err := s.listenConns[0].AcceptStream()
	if err != nil {
		return nil, err
	}
	log.Debugf("Dialed In")

	s.listenConns[0].SetStream(stream)

	select {
	case s.listenConns[0].Ready <- true:
	default:
	}

	log.Debugf("Set connection ready")

	// TODO: Rethink this
	/*go func(listenConn *packets.QUICReliableConn) {
		for {
			log.Debugf("Accepting new Stream on listen socket")
			stream, err := listenConn.AcceptStream()
			if err != nil {
				log.Fatalf("QUIC Accept err %s", err.Error())
			}

			log.Debugf("Accepted new Stream on listen socket")

			newConn := &packets.QUICReliableConn{}
			newConn.SetLocal(*s.localAddr)
			newConn.SetStream(stream)

			s.listenConns = append(s.listenConns, newConn)
		}
	}(s.listenConns[0])*/

	_, err = stream.Read(bts)
	if err != nil {
		return nil, err
	}
	p := DialPacketQuic{}
	network := bytes.NewBuffer(bts) // Stand-in for a network connection
	dec := gob.NewDecoder(network)
	err = dec.Decode(&p)
	if err != nil {
		return nil, err
	}

	log.Debugf("Waiting for %d more connections", p.NumPaths-1)

	for i := 1; i < p.NumPaths; i++ {
		log.Debugf("Got into loop for %d and %d", i, p.NumPaths)
		_, err := s.WaitForIncomingConn()
		log.Debugf("Having incoming conn")
		if err != nil {
			return nil, err
		}
		log.Debugf("Dialed In %d of %d", i, p.NumPaths)
	}

	// s.listenConns[0].SetPath(&p.Path)
	// log.Debugf("Got path from connection %v", p.Path)
	addr := p.Addr
	return &addr, nil
}

func (s *QUICSocket) Dial(remote snet.UDPAddr, path snet.Path, options DialOptions, i int) (packets.UDPConn, error) {
	// appnet.SetPath(&remote, path)
	if s.options == nil || !s.options.MultiportMode {
		conn := &packets.QUICReliableConn{}
		conn.SetLocal(*s.localAddr)
		err := conn.Dial(remote, &path)
		if err != nil {
			return nil, err
		}

		log.Debugf("Sending addr packet %d for conn %p", options.SendAddrPacket, &conn)
		if options.SendAddrPacket {
			var network bytes.Buffer
			enc := gob.NewEncoder(&network) // Will write to network.
			p := DialPacketQuic{
				Addr:     *s.localAddr,
				NumPaths: options.NumPaths,
			}

			err := enc.Encode(p)
			conn.Write(network.Bytes())
			if err != nil {
				return nil, err
			}
		}

		s.dialConns = append(s.dialConns, conn)

		return conn, nil
	} else {
		conn := &packets.QUICReliableConn{}
		conn.SetLocal(*s.localAddr)
		rem := remote.Copy()
		rem.Host.Port = remote.Host.Port + i
		log.Debugf("Remote port is %d", rem.Host.Port)
		err := conn.Dial(*rem, &path)
		if err != nil {
			return nil, err
		}

		log.Debugf("Sending addr packet %d for conn %p", options.SendAddrPacket, &conn)
		if options.SendAddrPacket {
			var network bytes.Buffer
			enc := gob.NewEncoder(&network) // Will write to network.
			p := DialPacketQuic{
				Addr:     *s.localAddr,
				NumPaths: options.NumPaths,
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
}

func (s *QUICSocket) DialAll(remote snet.UDPAddr, path []snet.Path, options DialOptions) ([]packets.UDPConn, error) {
	// TODO: Rethink this

	/*go func(listenConn *packets.QUICReliableConn) {

		stream, err := listenConn.AcceptStream()
		if err != nil {
			log.Fatalf("QUIC Accept err %s", err.Error())
		}
		s.listenConns[0].SetStream(stream)

		select {
		case s.listenConns[0].Ready <- true:
		default:
		}

		for {
			log.Debugf("Accepting new Stream on listen socket")
			stream, err := listenConn.AcceptStream()
			if err != nil {
				log.Fatalf("QUIC Accept err %s", err.Error())
			}

			log.Debugf("Accepted new Stream on listen socket")

			newConn := &packets.QUICReliableConn{}
			newConn.SetLocal(*s.localAddr)
			newConn.SetStream(stream)

			s.listenConns = append(s.listenConns, newConn)
		}
	}(s.listenConns[0])*/

	if options.NumPaths == 0 && len(path) > 0 {
		options.NumPaths = len(path)
	}

	// TODO: Differentiate between client/server based selection
	conns := make([]packets.UDPConn, 0)
	// conns[0] = s.listenConns[0]
	for i, v := range path {
		conn, err := s.Dial(remote, v, options, i)
		if err != nil {
			return nil, err
		}
		conns = append(conns, conn)
		time.Sleep(1 * time.Second)
	}

	select {
	case s.listenConns[0].Ready <- true:
		log.Debugf("Set Connection Ready")
	default:
		// s.listenConns[0]
	}

	return conns, nil
}

func (s *QUICSocket) GetConnections() []packets.UDPConn {
	conns := make([]packets.UDPConn, 0)
	for _, v := range s.listenConns {
		conns = append(conns, v)
	}
	for _, v := range s.dialConns {
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
