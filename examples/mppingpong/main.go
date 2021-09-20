package main

import (
	"encoding/gob"
	"flag"
	smp "github.com/netsys-lab/scion-path-discovery/api"
	"github.com/netsys-lab/scion-path-discovery/packets"
	"github.com/netsys-lab/scion-path-discovery/pathselection"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/lib/snet/path"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
	"time"
)

//LastSelection users could add more fields
type LastSelection struct {
	lastSelectedPathSet pathselection.PathSet
}

//CustomPathSelectAlg this is where the user actually wants to implement its logic in
func (lastSel *LastSelection) CustomPathSelectAlg(pathSet *pathselection.PathSet) (*pathselection.PathSet, error) {
	return pathSet.GetPathHighBandwidth(*numConns), nil
}

var numConns *int = flag.Int("n", 1, "Max number of outgoing connections")
var localAddr *string = flag.String("l", "localhost:9999", "Set the local address")
var remoteAddr *string = flag.String("r", "localhost:80", "Set the remote address")
var isServer *bool = flag.Bool("s", false, "Run as Server (otherwise, client)")
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

func main() {
	// peers := []string{"peer1", "peer2", "peer3"} // Later real addresses
	flag.Parse()
	setLoging()
	log.SetLevel(log.DebugLevel)
	gob.Register(path.Path{})
	lastSelection := LastSelection{}

	mpSock := smp.NewMPPeerSock(*localAddr, nil)
	err := mpSock.Listen()
	if err != nil {
		log.Fatal("Failed to listen MPPeerSock", err)
		os.Exit(1)
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

	if *isServer {
		remote, err := mpSock.WaitForPeerConnect(&lastSelection)
		if err != nil {
			log.Fatal("Failed to connect in-dialing peer", err)
		}
		log.Infof("Connected to %s", remote.String())
		for {
			for _, conn := range mpSock.UnderlaySocket.GetConnections() {
				bts := make([]byte, packets.PACKET_SIZE)
				if conn.GetType() == packets.ConnectionTypes.Incoming {
					n, err := mpSock.Read(bts)
					connNum := string(bts[:n])
					//var p snet.Path
					//network := bytes.NewBuffer(bts) // Stand-in for a network connection
					//dec := gob.NewDecoder(network)
					//err = dec.Decode(&p)
					//
					//println(p, i)

					if err != nil {
						log.Fatalf("Failed to read bytes from peer %s, err: %v", remote.String(), err)
					}
					log.Debugf("Read %d bytes from %s (origin %s)", n, remote.String(), connNum)
					log.Infof("Ping from %s", remote.String())
				}
				//if conn.GetType() == packets.ConnectionTypes.Outgoing {
				//	n, err := mpSock.Write(bts)
				//	if err != nil {
				//		log.Fatalf("Failed to write bytes from peer %s, err: %v", remote.String(), err)
				//	}
				//	log.Debugf("Wrote %d bytes to %s over conn %d", n, remote.String(), i)
				//	log.Infof("Pong to %s", remote.String())
				//}
				//time.Sleep(1 * time.Second)
			}
		}







	} else {
		peerAddr, err := snet.ParseUDPAddr(*remoteAddr)
		if err != nil {
			log.Fatalf("Failed to parse remote addr %s, err: %v", *remoteAddr, err)
			os.Exit(1)
		}
		mpSock.SetPeer(peerAddr)

		err = mpSock.Connect(&lastSelection, nil)
		log.Infof("Connected to %s", *remoteAddr)
		if err != nil {
			log.Fatal("Failed to connect MPPeerSock", err)
		}

		for {
			for i, conn := range mpSock.UnderlaySocket.GetConnections() {
				//bts := make([]byte, packets.PACKET_SIZE)
				if conn.GetType() == packets.ConnectionTypes.Outgoing {
					//var network bytes.Buffer
					//enc := gob.NewEncoder(&network) // Will write to network.
					//err := enc.Encode(conn.GetPath())
					n, err := mpSock.Write([]byte(strconv.Itoa(i)))
					if err != nil {
						log.Fatalf("Failed to write bytes from peer %s, err: %v", *remoteAddr, err)
					}
					log.Debugf("Wrote %d bytes to %s over conn %d", n, *remoteAddr, i)
					log.Infof("Ping to %s", *remoteAddr)
				}
				//if conn.GetType() == packets.ConnectionTypes.Incoming {
				//	n, err := mpSock.Read(bts)
				//	if err != nil {
				//		log.Fatalf("Failed to read bytes from peer %s, err: %v", *remoteAddr, err)
				//	}
				//	log.Debugf("Read %d bytes from %s", n, *remoteAddr)
				//	log.Infof("Pong from %s", *remoteAddr)
				//}
				time.Sleep(1 * time.Second)
			}
		}
	}

	// mpSock.SetPeer(remote)
	// mpSock.Connect(customPathSelectAlg)
	defer mpSock.Disconnect()

}
