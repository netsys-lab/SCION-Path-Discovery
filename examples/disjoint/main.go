package main

import (
	"flag"
	"log"
	"os"
	"sync"
	"time"

	"github.com/netsec-ethz/scion-apps/pkg/pan"
	smp "github.com/netsys-lab/scion-path-discovery/api"
	"github.com/netsys-lab/scion-path-discovery/packets"
	"github.com/netsys-lab/scion-path-discovery/pathselection"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/sirupsen/logrus"
)

var localAddr *string = flag.String("l", "localhost:9999", "Set the local address")
var remoteAddr *string = flag.String("r", "18-ffaa:1:ef8,[127.0.0.1]:12345", "Set the remote address")
var isServer *bool = flag.Bool("s", false, "Run as Server (otherwise, client)")

func main() {
	// peers := []string{"peer1", "peer2", "peer3"} // Later real addresses
	flag.Parse()
	logrus.SetLevel(logrus.DebugLevel)
	mpSock := smp.NewPanSock(*localAddr, nil, &smp.PanSocketOptions{
		Transport: "QUIC",
	})
	err := mpSock.Listen()
	if err != nil {
		log.Fatal("Failed to listen PanSock: ", err)
		os.Exit(1)
	}

	if *isServer {

		i := 1

		for {
			remote, err := mpSock.WaitForPeerConnect()
			if err != nil {
				log.Fatal("Failed to connect in-dialing peer: ", err)
				os.Exit(1)
			}
			log.Printf("Connected to %s", remote.String())
			addr, _ := pan.ResolveUDPAddr(*localAddr)
			addr.Port = addr.Port + uint16(i)*32

			mps := smp.NewPanSock(addr.String(), nil, &smp.PanSocketOptions{
				Transport: "QUIC",
			})

			err = mps.Listen()
			if err != nil {
				log.Fatal("Failed to listen mps: ", err)
				os.Exit(1)
			}

			mps.SetPeer(remote)
			disjointSel := smp.NewDisjointPathSelectionSocket(mps, 2, 2)
			go func(dj *smp.DisjointPathselection) {
				metricsTicker := time.NewTicker(1 * time.Second)
				for {
					select {
					case <-metricsTicker.C:
						_, err := disjointSel.UpdatePathSelection()
						if err != nil {
							logrus.Error("[DisjointPathSelection] Failed to update path selection ", err)
							os.Exit(1)
						}
					}
				}

			}(disjointSel)

			pathset, err := disjointSel.InitialPathset()
			if err != nil {
				log.Fatal("Failed to obtain pathset", err)
				os.Exit(1)
			}
			pathset.Address = *remote

			err = mps.Connect(&pathset, nil)
			if err != nil {
				log.Fatal("Failed to connect MPPeerSock", err)
				os.Exit(1)
			}

			i++
			go func(mps *smp.PanSocket) {
				WriteAllConns(mps)
				logrus.Warn("Done writing all conns to ", mps.Peer.String())
			}(mps)
		}

	} else {
		peerAddr, err := snet.ParseUDPAddr(*remoteAddr)
		if err != nil {
			log.Fatalf("Failed to parse remote addr %s, err: %v", *remoteAddr, err)
			os.Exit(1)
		}

		mpSock.SetPeer(peerAddr)
		paths, _ := mpSock.GetAvailablePaths()
		pathset := pathselection.WrapPathset(paths)
		pathset.Address = *peerAddr

		err = mpSock.Connect(&pathset, nil)
		if err != nil {
			log.Fatal("Failed to connect MPPeerSock", err)
			os.Exit(1)
		}

		for {
			mpSock.Disconnect()
			new, err := mpSock.WaitForPeerConnect()
			if err != nil {
				log.Fatal("Failed to wait for back MPPeerSock", err)
				os.Exit(1)
			}
			logrus.Info("Got conn back from remote ", new.String())
			ReadAllConns(mpSock)
		}
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
				_, err := c.Read(bts)
				if err != nil {
					logrus.Error(err)
					wg.Done()
				}
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
				_, err := c.Write(bts)
				if err != nil {
					logrus.Error(err)
					wg.Done()
				}
			}
		}(c)
	}

	wg.Wait()
}
