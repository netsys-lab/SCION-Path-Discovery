package main

import (
	"bytes"
	"encoding/gob"
	"flag"
	"sync"
	"time"

	smp "github.com/netsys-lab/scion-path-discovery/api"
	"github.com/netsys-lab/scion-path-discovery/packets"
	lookup "github.com/netsys-lab/scion-path-discovery/pathlookup"
	"github.com/netsys-lab/scion-path-discovery/pathselection"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

type PathPacket struct {
	Path string
}

var numConns *int = flag.Int("n", 1, "Max number of outgoing connections")
var localAddr *string = flag.String("l", "", "Set the local address")
var remoteAddr *string = flag.String("r", "", "Set the remote address")
var transport *string = flag.String("t", "SCION", "Set the transprt (SCION|QUIC)")
var loglevel *string = flag.String("loglevel", "INFO", "TRACE|DEBUG|INFO|WARN|ERROR|FATAL")

func setLogging() {
	if loglevel == nil {
		return
	}

	switch *loglevel {
	case "TRACE":
		log.SetLevel(log.TraceLevel)
		break
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
		break
	case "INFO":
		log.SetLevel(log.InfoLevel)
		break
	case "WARN":
		log.SetLevel(log.WarnLevel)
		break
	case "ERROR":
		log.SetLevel(log.ErrorLevel)
		break
	case "FATAL":
		log.SetLevel(log.FatalLevel)
		break
	}
}

func ReadAllConns(mps *smp.PanSocket) {
	var wg sync.WaitGroup
	conns := mps.UnderlaySocket.GetConnections()
	for _, c := range conns {
		wg.Add(1)
		go func(c packets.UDPConn) {
			bts := make([]byte, packets.PACKET_SIZE)
			for {
				n, err := c.Read(bts)
				if err != nil {
					logrus.Errorf("Failed to read bytes from peer %s, err: %v", *remoteAddr, err)
					wg.Done()
				}

				pkt := PathPacket{}
				network := bytes.NewBuffer(bts)
				dec := gob.NewDecoder(network)
				err = dec.Decode(&pkt)

				log.Debugf("Read %d bytes from %s", n, mps.Peer.String())
				log.Infof("Ping from %s over %s", mps.Peer.String(), pkt.Path)
			}
		}(c)
	}

	wg.Wait()
}

func WriteAllConns(mps *smp.PanSocket) {
	var wg sync.WaitGroup
	conns := mps.UnderlaySocket.GetConnections()
	for _, c := range conns {
		wg.Add(1)
		go func(c packets.UDPConn) {
			// bts := make([]byte, packets.PACKET_SIZE)
			for {

				pkt := PathPacket{}
				path := c.GetPath()
				log.Error(*path)
				p := lookup.PathToString(*path)
				pkt.Path = p
				var network2 bytes.Buffer
				enc := gob.NewEncoder(&network2)
				err := enc.Encode(pkt)
				log.Error(pkt)
				if err != nil {
					log.Fatal(err)
				}
				n, err := c.Write(network2.Bytes())
				if err != nil {
					logrus.Errorf("Failed to read bytes from peer %s, err: %v", *remoteAddr, err)
					wg.Done()
				}

				log.Debugf("Wrote %d bytes fto %s", n, *remoteAddr)
				log.Infof("Ping to %s over %s", *remoteAddr, pkt.Path)
				time.Sleep(1 * time.Second)
			}
		}(c)
	}

	wg.Wait()
}

func main() {
	flag.Parse()
	gob.Register(PathPacket{})
	setLogging()

	mpSock := smp.NewPanSock(*localAddr, nil, &smp.PanSocketOptions{
		Transport: *transport,
	})
	err := mpSock.Listen()
	if err != nil {
		log.Fatal("Failed to listen MPPeerSock", err)
	}

	log.Infof("Listening on %s", *localAddr)
	if remoteAddr == nil || *remoteAddr == "" {
		// TODO: Remote and Get Path is not working anymore
		remote, err := mpSock.WaitForPeerConnect()
		if err != nil {
			log.Fatal("Failed to wait for MPPeerSock connect", err)
		}
		mpSock.SetPeer(remote)
		ReadAllConns(mpSock)
	} else {
		peerAddr, err := snet.ParseUDPAddr(*remoteAddr)
		if err != nil {
			log.Fatalf("Failed to parse remote addr %s, err: %v", *remoteAddr, err)
		}
		mpSock.SetPeer(peerAddr)
		paths, err := mpSock.GetAvailablePaths()
		if err != nil {
			log.Fatal("Failed to fetch paths to ", peerAddr, " err: ", err)
		}
		paths = paths[:*numConns]
		ps := pathselection.WrapPathset(paths)
		ps.Address = *peerAddr
		err = mpSock.Connect(&ps, nil)
		log.Infof("Connected to %s", *remoteAddr)
		if err != nil {
			log.Fatal("Failed to connect MPPeerSock", err)
		}
		WriteAllConns(mpSock)
	}

	if remoteAddr == nil || *remoteAddr == "" {

	} else {

	}

	defer mpSock.Disconnect()
}
