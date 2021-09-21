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
				if conn.GetType() == packets.ConnectionTypes.Incoming {
					bts := make([]byte, packets.PACKET_SIZE)
					//n, err := mpSock.Read(bts)
					n, err := conn.Read(bts)
					if err != nil {
						log.Fatalf("Failed to read bytes from peer %s, err: %v", remote.String(), err)
					}
					connNum := string(bts[:n])


					if n > 1 {continue}


					log.Debugf("Read %d bytes from %s (origin %s)", n, remote.String(), connNum)
					log.Infof("Ping from %s", remote.String())
					n, err = mpSock.Write(bts)
					//n, err = conn.Write(bts)
					if err != nil {
						log.Fatalf("Failed to write bytes from peer %s, err: %v", remote.String(), err)
					}
					log.Debugf("Wrote %d bytes to %s", n, remote.String())
					log.Infof("Pong to %s", remote.String())
				}
				//time.Sleep(1 * time.Second)
			}
		}







	} else {
		peerAddr, err := snet.ParseUDPAddr(*remoteAddr)
		if err != nil {
			log.Fatalf("Failed to parse remote addr %s, err: %v", *remoteAddr, err)
		}
		mpSock.SetPeer(peerAddr)

		err = mpSock.Connect(&lastSelection, nil)
		log.Infof("Connected to %s", *remoteAddr)
		if err != nil {
			log.Fatal("Failed to connect MPPeerSock", err)
		}

		for {
			for i, conn := range mpSock.UnderlaySocket.GetConnections() {
				if conn.GetType() == packets.ConnectionTypes.Outgoing {
					bts := make([]byte, packets.PACKET_SIZE)
					//n, err := mpSock.Write([]byte(strconv.Itoa(i)))
					n, err := conn.Write([]byte(strconv.Itoa(i)))
					if err != nil {
						log.Fatalf("Failed to write bytes from peer %s, err: %v", *remoteAddr, err)
					}
					log.Debugf("Wrote %d bytes to %s over conn %d", n, *remoteAddr, i)
					log.Infof("Ping to %s", *remoteAddr)
					n, err = mpSock.Read(bts)
					//n, err = conn.Read(bts)
					if err != nil {
						log.Fatalf("Failed to read bytes from peer %s, err: %v", *remoteAddr, err)
					}
					connNum, err := strconv.Atoi(string(bts[:1]))
					println(i, string(bts[:n]), connNum)
					if err != nil || connNum != i {
						log.Infof("Pong from wrong connection %d", connNum)
					}
					log.Debugf("Read %d bytes from %s", n, *remoteAddr)
					log.Infof("Pong from %s", *remoteAddr)
				}
				time.Sleep(1 * time.Second)
			}
		}
	}

	// mpSock.SetPeer(remote)
	// mpSock.Connect(customPathSelectAlg)
	//defer mpSock.Disconnect()

}
