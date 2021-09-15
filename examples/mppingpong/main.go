package main

import (
	"flag"
	"os"
	"time"

	smp "github.com/netsys-lab/scion-path-discovery/api"
	"github.com/netsys-lab/scion-path-discovery/packets"
	"github.com/netsys-lab/scion-path-discovery/pathselection"
	"github.com/scionproto/scion/go/lib/snet"
	log "github.com/sirupsen/logrus"
)

//LastSelection users could add more fields
type LastSelection struct {
	lastSelectedPathSet pathselection.PathSet
}

//CustomPathSelectAlg this is where the user actually wants to implement its logic in
func (lastSel *LastSelection) CustomPathSelectAlg(pathSet *pathselection.PathSet) (*pathselection.PathSet, error) {
	return pathSet.GetPathHighBandwidth(3), nil
}

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
			conns := <-mpSock.OnConnectionsChange
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
			os.Exit(1)
		}
		log.Infof("Connected to %s", remote.String())
		for {
			bts := make([]byte, packets.PACKET_SIZE)
			n, err := mpSock.Read(bts)
			if err != nil {
				log.Fatalf("Failed to read bytes from peer %s, err: %v", remote.String(), err)
				os.Exit(1)
			}
			log.Debugf("Read %d bytes from %s", n, remote.String())
			log.Infof("Pong from %s", remote.String())
			n, err = mpSock.Write(bts)
			if err != nil {
				log.Fatalf("Failed to write bytes from peer %s, err: %v", remote.String(), err)
				os.Exit(1)
			}
			log.Debugf("Wrote %d bytes to %s", n, remote.String())
			log.Infof("Ping to %s", remote.String())
			time.Sleep(1 * time.Second)
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
			os.Exit(1)
		}
		for {
			bts := make([]byte, packets.PACKET_SIZE)
			n, err := mpSock.Write(bts)
			if err != nil {
				log.Fatalf("Failed to write bytes from peer %s, err: %v", *remoteAddr, err)
				os.Exit(1)
			}
			log.Debugf("Wrote %d bytes to %s", n, *remoteAddr)
			log.Infof("Pong to %s", *remoteAddr)
			n, err = mpSock.Read(bts)
			if err != nil {
				log.Fatalf("Failed to read bytes from peer %s, err: %v", *remoteAddr, err)
				os.Exit(1)
			}
			log.Debugf("Read %d bytes from %s", n, *remoteAddr)
			log.Infof("Ping from %s", *remoteAddr)
			time.Sleep(1 * time.Second)
		}

	}

	// mpSock.
	// mpSock.SetPeer(remote)
	// mpSock.Connect(customPathSelectAlg)
	defer mpSock.Disconnect()

}
