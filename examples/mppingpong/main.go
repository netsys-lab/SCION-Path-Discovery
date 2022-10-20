package main

import (
	"bytes"
	"encoding/gob"
	"flag"
	"sync"
	"time"

	smp "github.com/netsys-lab/scion-path-discovery/api"
	"github.com/netsys-lab/scion-path-discovery/packets"
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

				log.Debugf("Read %d bytes from %s", n, *remoteAddr)
				log.Infof("Ping from %s over %s", *remoteAddr, pkt.Path)
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
			bts := make([]byte, packets.PACKET_SIZE)
			for {
				n, err := c.Write(bts)
				if err != nil {
					logrus.Errorf("Failed to read bytes from peer %s, err: %v", *remoteAddr, err)
					wg.Done()
				}

				pkt := PathPacket{}
				network := bytes.NewBuffer(bts)
				dec := gob.NewDecoder(network)
				err = dec.Decode(&pkt)

				log.Debugf("Read %d bytes from %s", n, *remoteAddr)
				log.Infof("Ping from %s over %s", *remoteAddr, pkt.Path)
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

	peerAddr, err := snet.ParseUDPAddr(*remoteAddr)
	if err != nil {
		log.Fatalf("Failed to parse remote addr %s, err: %v", *remoteAddr, err)
	}
	mpSock := smp.NewPanSock(*localAddr, peerAddr, nil)
	err = mpSock.Listen()
	if err != nil {
		log.Fatal("Failed to listen MPPeerSock", err)
	}

	log.Infof("Listening on %s", *localAddr)
	paths, err := mpSock.GetAvailablePaths()
	if err != nil {
		log.Fatal("Failed to fetch paths to ", peerAddr, " err: ", err)
	}
	paths = paths[:*numConns]
	ps := pathselection.WrapPathset(paths)
	err = mpSock.Connect(&ps, nil)
	log.Infof("Connected to %s", *remoteAddr)
	if err != nil {
		log.Fatal("Failed to connect MPPeerSock", err)
	}

	if remoteAddr == nil || *remoteAddr == "" {
		ReadAllConns(mpSock)
	} else {
		WriteAllConns(mpSock)
	}

	defer mpSock.Disconnect()
}
