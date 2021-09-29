package main

import (
	"bytes"
	"encoding/gob"
	"flag"
	"fmt"
	smp "github.com/netsys-lab/scion-path-discovery/api"
	"github.com/netsys-lab/scion-path-discovery/packets"
	"github.com/netsys-lab/scion-path-discovery/pathselection"
	"github.com/scionproto/scion/go/lib/snet"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

type PathPacket struct {
	Path string
}

//LastSelection users could add more fields
type LastSelection struct {
	lastSelectedPathSet pathselection.PathSet
}

//CustomPathSelectAlg this is where the user actually wants to implement its logic in
func (lastSel *LastSelection) CustomPathSelectAlg(pathSet *pathselection.PathSet) (*pathselection.PathSet, error) {
	return pathSet.GetPathSmallHopCount(*numConns), nil
}

var numConns *int = flag.Int("n", 1, "Max number of outgoing connections")
var localAddr *string = flag.String("l", "localhost:9999", "Set the local address")
var remoteAddr *string = flag.String("r", "localhost:80", "Set the remote address")
var loglevel *string = flag.String("loglevel", "INFO", "TRACE|DEBUG|INFO|WARN|ERROR|FATAL")

func setLoging() {
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

func PathToString(path snet.Path) string {
	if path == nil {
		return ""
	}
	intfs := path.Metadata().Interfaces
	if len(intfs) == 0 {
		return ""
	}
	var hops []string
	intf := intfs[0]
	hops = append(hops, fmt.Sprintf("%s %s",
		intf.IA,
		intf.ID,
	))
	for i := 1; i < len(intfs)-1; i += 2 {
		inIntf := intfs[i]
		outIntf := intfs[i+1]
		hops = append(hops, fmt.Sprintf("%s %s %s",
			inIntf.ID,
			inIntf.IA,
			outIntf.ID,
		))
	}
	intf = intfs[len(intfs)-1]
	hops = append(hops, fmt.Sprintf("%s %s",
		intf.ID,
		intf.IA,
	))
	return fmt.Sprintf("[%s]", strings.Join(hops, ">"))
}

func receivePackets(mpSock *smp.MPPeerSock) {
	go func() {
		//search for incoming connection. There is only one (until now).
		for _, conn := range mpSock.UnderlaySocket.GetConnections() {
			//stay at incoming connection.
			for {
				if conn.GetType() == packets.ConnectionTypes.Incoming {
					bts := make([]byte, packets.PACKET_SIZE)
					n, err := conn.Read(bts)
					if err != nil {
						log.Fatalf("Failed to read bytes from peer %s, err: %v", *remoteAddr, err)
					}

					pkt := PathPacket{}
					network := bytes.NewBuffer(bts)
					dec := gob.NewDecoder(network)
					err = dec.Decode(&pkt)

					log.Debugf("Read %d bytes from %s", n, *remoteAddr)
					log.Infof("Ping from %s over %s", *remoteAddr, pkt.Path)
				}
			}
		}
	}()
}

func sendPackets(mpSock *smp.MPPeerSock) {
	go func() {
		for {
			for _, conn := range mpSock.UnderlaySocket.GetConnections() {
				if conn.GetType() == packets.ConnectionTypes.Outgoing {
					str := PathToString((*conn.GetPath()).Copy())

					var network bytes.Buffer
					enc := gob.NewEncoder(&network)
					pkt := PathPacket{
						Path: str,
					}
					err := enc.Encode(&pkt)
					n, err := conn.Write(network.Bytes())

					if err != nil {
						log.Fatalf("Failed to write bytes from peer %s, err: %v", *remoteAddr, err)
					}
					log.Debugf("Wrote %d bytes to %s", n, *remoteAddr)
					log.Infof("Ping to %s over %s", *remoteAddr, str)
				}
			}
			time.Sleep(1 * time.Second)
		}
	}()
}

func main() {
	flag.Parse()
	setLoging()
	gob.Register(PathPacket{})
	lastSelection := LastSelection{}

	peerAddr, err := snet.ParseUDPAddr(*remoteAddr)
	if err != nil {
		log.Fatalf("Failed to parse remote addr %s, err: %v", *remoteAddr, err)
	}
	mpSock := smp.NewMPPeerSock(*localAddr, peerAddr)
	err = mpSock.Listen()
	if err != nil {
		log.Fatal("Failed to listen MPPeerSock", err)
	}

	go func() {
		for {
			log.Info("Waiting for new connections")
			conns := <- mpSock.OnConnectionsChange
			log.Infof("New Connections available, got %d", len(conns))
			for i, v := range conns {
				log.Infof("Connection %d is %s", i, packets.ConnTypeToString(v.GetType()))
			}
		}
	}()

	log.Infof("Listening on %s", *localAddr)

	err = mpSock.Connect(&lastSelection, nil)
	log.Infof("Connected to %s", *remoteAddr)
	if err != nil {
		log.Fatal("Failed to connect MPPeerSock", err)
	}

	sendPackets(mpSock)
	receivePackets(mpSock)

	defer mpSock.Disconnect()

	select {}
}
