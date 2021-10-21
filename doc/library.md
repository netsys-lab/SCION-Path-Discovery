# Multipath Library - Implementation

## Overview

The following components are compound to implement the multipath library
- **UnderlaySocket**: Provides listen and connect methods that return TransportConnections
- **Connection**: Abstracts underlay connections providing metrics and read/write methods
- **PathQuality**: Stores metrics and meta information about paths
- **PathQualityDB**: Serves as database for PathQuality entries
- **CustomPathSelection**: Provides an interface to implement custom path selection algorithms
- **MPPeerSock**: Represents a multipath socket to a particular peer

The following Figure illustrates, how these compontens are put together to form a working multipath library.

![pathdiscovery-abstractions (1)](https://user-images.githubusercontent.com/32448709/137099751-ec4233a6-6312-407b-ab94-1139c484029b.jpg)


The application starts the multipath socket by creating a new instance of MPPeerSock and passes a custom pathselection algorithm optionally. To start communication, the Listen method is called by the application. Afterwards, the MPPeerSock can either call Connect to another listening socket or call Accept to wait for incoming connections. These connections are instantiated with Connection components. Each Connection runs over a particular path. Data can be read or written using Read and Write methods of one or more connections, concurrently. These connections collect metrics for used paths which are fetched from the MPPeerSock instance and stored in the PathQualityDB. Using a timer, the path selection is performed repeatedly after a configured amount of time. If a CustomPathSelection is passed to the MPPeerSock, this component will be called with the paths and all collected metrics combined in PathQuality entries and returns a subset of these paths. The selected subset will then be passed to the UnderlaySocket and the PacketScheduler to adapt the list of Connections that are used.

## Pathselection
To implement pathselection for particular apps, this library defines an interface that needs to be implemented, called CustomPathSelection.

```go
type CustomPathSelection interface {
	CustomPathSelectAlg(*PathSet) (*PathSet, error)
}
```

This interface needs to instantiated and passed to the MPPeerSock. Everytime new PathQualities are available, the `CustomPathSelectAlg` is executed, getting the new PathSet containing all PathQualities. The algorithm now needs to decide, out of all available PathQualities, which paths should be used. It returns a new pathset, containing all selected PathQualities, or an error if something unexpected happened.

The MPPeerSock now handles this selection result and ensures, that connections over the selected paths are open. Connections over already used paths are kept untouched, new paths will be handled by creating new connections. Connections over unselected paths are marked as closed, so that the application can react to stop using them and close them gracefully.

## Connection Types
At the moment, we support two different Connections/UnderlaySockets: SCION over plain UDP (SCION/UDP) and SCION over QUIC (SCION/QUIC). SCION/UDP does not support bidirectional connections, meaning dialing to a remote peer does not create a stateful connection between the peers. Furthermore, there is no guarantee that packets are transferred properly. Finally, both peers may use a different number of outgoing connections to the particular remote peer, since all incoming SCION packets arrive at the same listening connection. The next Figure shows how communication using SCION/UDP looks:

![pathdisc](https://user-images.githubusercontent.com/32448709/137102316-0c98273c-40f1-4399-9f25-60ae8da27f23.jpg)

SCION/QUIC, using quic-go as QUIC implementation underneath, works with stateful connections. This means both peers need to agree on a number of bidirectional connections. For SCION/QUIC, SCION packets sent over a particular connection arrive need to be read from the remote peer at exactly that connection, compared to SCION/UDP where all packets arrive at the same connection. The next Figure shows how SICON/QUIC looks:

![pathdisc (1)](https://user-images.githubusercontent.com/32448709/137102881-b6d56d0a-84ac-4dc0-b9d3-2cea9c615333.jpg)


## Using Multiple Paths
After connecting to a peer using the `Connect` method, a slice of connections is can be fetched via `sock.UnderlaySocket.GetConnections`, where each connection uses one of the selected path internally. The library provides the channel `OnConnectionsChange` that returns new connections each time an internal ticker starts performing new pathselection. Each connection has a `GetId` method which returns its unique identifier (and also the one of its underlying path), so applications can check if the returned connections changed. The library does not dial again over already used paths. A sample of using those methods looks like this:

```go
for _, conn := range mpSock.UnderlaySocket.GetConnections() {
 // ...
}

for {
	log.Info("Waiting for new connections")
	conns := <-mpSock.OnConnectionsChange
	log.Infof("New Connections available, got %d", len(conns))
	for i, v := range conns {
		var str string = ""
		path := v.GetPath()
		if path != nil {
			str = PathToString(*path)
		}

		log.Infof("Connection %d is %s, path %s", i, packets.ConnTypeToString(v.GetType()), str)
	}
}
```

## Extensibility
We aim the design of this library to be easily extensible for further metrics, Connections or UnderlaySockets. New UnderlaySockets and/or Connections can be added without touching the existing ones and may be added via the socketOptions "Transport" flag. An UnderlaySocket may also be extended to use different Connections, e.g. the [snet](https://github.com/scionproto/scion/tree/master/go/lib/snet) SCION Connection or the [SCION OptimizedConn](https://github.com/johannwagner/scion-optimized-connection). By introducing the CustomPathSelection interface, applications can easily implement different kinds of pathselection without the need of touching the library, but with helpful utilities to pre-sort paths.

## Example: Multipath PingPong
To test the multipath capabilities of this library, we provide an example, called multipath pingpong. This example can be started with a local, a remote SCION address and a number of outgoing connections n. We call one running instance of this example a peer. To see how multipath communication works, two peers need to be started. Each peer sends ping packets over n connections and reads all incoming pong connections, printing over which paths the pings are sent and over which paths the pongs are received. This example is using SCION/UDP connections, meaning there is no handshake to initiate bidiretional connections. 

Host 1: `./mppingpong -l "19-ffaa:1:c3f,[141.44.25.148]:32000" -r "19-ffaa:1:cf0,[141.44.25.151]:32000"  -n 2`

Host 2: `./mppingpong -r 19-ffaa:1:c3f,141.44.25.148:32000 -l 19-ffaa:1:cf0,141.44.25.151:32000  -n 3 `
