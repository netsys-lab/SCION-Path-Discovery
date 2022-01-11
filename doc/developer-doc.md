# SCION Multipath Library - Developer Guide

This documentation is targeted for developers that want to write SCION application based on the SCION Multipath Library (smp). 

**Further resources**:
- [Documetation of concept](https://github.com/netsys-lab/scion-path-discovery/blob/main/doc/path-selection.org)
- [Documentation of implementation](https://github.com/netsys-lab/scion-path-discovery/blob/main/doc/library.md)
- [Go doc](https://pkg.go.dev/github.com/netsys-lab/scion-path-discovery)

## Complete Examples
Before diving into explaining particular functionalities of the library, for an easy start, please refer to the existing examples that already implement smp:
- [MPPingPong](https://github.com/netsys-lab/scion-path-discovery/blob/main/examples/mppingpong/main.go): A program that sends and receives ping and pong messages to other MPPingPong instances using up to `n` paths
- [BitTorrent over SCION](https://github.com/netsys-lab/bittorrent-over-scion): Our BitTorrent over SCION implementation that uses smp for multipath bandwidth aggregation

## Usage

### Creating Listening MPPeerSockets
To create a MPPeerSocket, initialize it via `smp.NewMPPeerSock` passing the local SCION address as a string to it. The second argument is the remote addr, which should be omitted for sockets that wait for incoming connections. Each instantiated socket must call `Listen`. Afterwards, socket that wait for incoming connections, call `WaitForPeerConnect`. Passing `nil` to this call means, that the peer that connects to this socket performs the path selection. Each MPPeerSock is designed to be connected to a single remote MPPeerSock, creating a 1:1 connection that allows using a variable number of paths.

```go
mpSock := smp.NewMPPeerSock(*localAddr, nil, nil)
err = mpSock.Listen()
if err != nil {
    log.Fatal("Failed to listen MPPeerSock", err)
}

// Wait for incoming connections
remoteAddr, err = mpSock.WaitForPeerConnect(nil)

if err != nil {
    return nil, err
}
```

The third argument provides options to configure the socket via `MPSocketOptions`. This argument may be omitted to apply to default configuration. Default Transport is "QUIC", MultiportMode is disabled by default.

```go
type MPSocketOptions struct {
    // Defines which underlying protocol should be used. 
	Transport                   string // "QUIC" | "SCION"
    PathSelectionResponsibility string // "CLIENT" | "SERVER" | "BOTH"
	// Multipoort lets each connection run over a dedicated port, which significantly improces performance
    MultiportMode               bool
}

```
### Connecting to MPPeerSockets
To connect to a remote MPPeerSocket, a new instance needs to be created and `Listen` must be called. Furthermore, the peer address needs to be passed to the the constructor, in the form of a `snet.UDPAddr`. After calling Listen, using the `Connect` method, a multipath connection to a waiting remote socket will be established. As first argument, an implementation of the `CustomPathSelection` interface needs to be passed to enable path selection. 

```go
peerAddr, err := snet.ParseUDPAddr(*remoteAddr)
if err != nil {
    log.Fatalf("Failed to parse remote addr %s, err: %v", *remoteAddr, err)
}
mpSock := smp.NewMPPeerSock(*localAddr, peerAddr, nil)
err = mpSock.Listen()
selection := CustomSelection{} // Implement CustomPathSelection interface
err = mpSock.Connect(&selection, nil)
```

The second argument to Connect are potential `ConnectOptions`. These should be omitted to enable default configuration. However, for fine tuning, the following options are possible:

```go
type ConnectOptions struct {
	SendAddrPacket          bool // Send a packet to the waiting socket containing local information
	DontWaitForIncoming     bool // Force Listening Socket not to wait for incoming connections
	NoPeriodicPathSelection bool // Disable automatic periodic path selection
}
```

### Changing the TransportType
Per default, TransportType `SCION` is used by the socket. To change this, change the `Transport` field in `MPSocketOptions`. At the moment, only `SCION` and `QUIC` are supported, further TransportTypes are planned to be integrated in the future.

### Handle multiple incoming MPPeerSockets
To handle multiple incoming MPPeerSocket instances on a particular port, the `MPListener` struct should be used. It provides the `WaitForMPPeerSockConnect`  method, that may be called after `Listen` in a loop. The method blocks until a new remote socket dialed in and it returns its `snet.UDPAddr`. This may be used to e.g. connect a new MPPeerSocket back. The listener avoids creating own MPPeerSocket instances in  WaitForMPPeerSockConnect (and instead returns only the actual address) to provide more flexibility.

```go
mpListener := smp.NewMPListener(s.lAddr, &smp.MPListenerOptions{
    Transport: "QUIC",
})

err = mpListener.Listen()
for {
    remote, err := mpListener.WaitForMPPeerSockConnect()
    // Here we could create own MPPeerSocks to Connect back, 
    // this is how BitTorrent over SCION does it's seeder-based path selection
}
```
The `MPListenerOptions` struct contains the underlying transport, which may be "SCION" or "QUIC", similar to `MPSocketOptions`.

### Using the instantiated Connections
After connecting to a peer using the `Connect` method, a slice of connections can be fetched via `sock.UnderlaySocket.GetConnections`, where each connection uses one of the selected paths internally. The library provides the channel `OnConnectionsChange` that returns new connections each time an internal ticker starts performing new pathselection. Each connection has a `GetId` method which returns its unique identifier (and also the one of its underlying path), so applications can check if the returned connections changed. The library does not dial again over already used paths. An example use of those methods is shown below:

```go
for _, conn := range mpSock.UnderlaySocket.GetConnections() {
 // ...
}

for {
	log.Info("Waiting for new connections")
	conns := <-mpSock.OnConnectionsChange
	
	}
}
```

### Configure Pathselection
Per default, the path selection (based on all up to date information), runs every 5 seconds. To write an own path selection algorithm, the This means the `CustomPathSelection` interface needs to be implemented and passed to `connect` and/or `waitForPeerConnect`. This interface has at the moment only one method, which gets a PathSet (containing paths and all their collected information) and returns again a PathSet, which will afterwards be applied to by the socket.

```go
type CustomPathSelection interface {
	CustomPathSelectAlg(*PathSet) (*PathSet, error)
}
```

As mentionend, `connect` and/or `waitForPeerConnect` expect a CustomPathSelection or nil to be passed. To implement such an interface, smp provides useful helper functions that do pre-sorting or filtering of paths. To filter for the `n` smallest hops, a CustomPathSelection could look like this:

```go
// e.g. a cmd flag
var numConns *int

type ShortedPathSelection struct {
	
}

func (sps *ShortedPathSelection) CustomPathSelectAlg(pathSet *pathselection.PathSet) (*pathselection.PathSet, error) {
    // Helper for getting up to numConns of the the shortest paths
	return pathSet.GetPathSmallHopCount(*numConns), nil
}
```

### Handle additional and closed connections
Using the `OnConnectionsChange` event, the application can wait for a new result of path selection. This event is fired every time the path selection was performed by the socket. 

```go
for {
	log.Info("Waiting for new connections")
	conns := <-mpSock.OnConnectionsChange
	
}
```

Your application should contain some kind of list of the currently used connections. Using this list, you can compare the connection list of the application with the new result of the socket using the `conn.GetId` method, which results a unique identifier of the connection. 

Basically, the application can do 4 different things with the results of the event:
1) No new connections (all are the same), do nothing.
2) New connections, meaning the selected pathset differs, add the connections to the application list.
3) A connection was replaced with one that got a new path, replace the connection in the application list.
4) Existing connections were marked as `closed` by the socket (this can be checked by `conn.GetState() == packets.ConnectionStates.Closed`), meaning the pathset does not contain the path of the connection anymore. Those connections should be closed gracefully by the application.

### Configure Logging
Smp uses [logrus](https://github.com/sirupsen/logrus) with Loglevel INFO per default. Configuring the loglevel or the log visualization, please refer to logrus documentation. A useful example configuration to display colored log messages with timestamps looks like this: 
```go
import (
	log "github.com/sirupsen/logrus"
)

log.SetFormatter(&log.TextFormatter{
    DisableColors: false,
    FullTimestamp: true,
})
log.SetLevel(log.DebugLevel)
```

### Incoming, outgoing and bidirectional connections
In [library.md](https://github.com/netsys-lab/scion-path-discovery/blob/main/doc/library.md#connection-types) we explained the two transport types: SCION and QUIC and how their connections behave. For SCION, there is always one incoming connection, and potential `n` outgoing connections. For two connected MPPeerSockets, the number of outoging connections, and consequently, the number of used paths for outgoing traffic, may differ (as can be seen in the MPPingPong example). Using QUIC as transport type, the connections are always bidirectional, meaning both connected MPPeerSocks have the same number of paths in use. We implement this using a `ConnectionTypes` enum:

```go
type connectionTypes struct {
	Incoming      int
	Outgoing      int
	Bidirectional int
}
```

To determine the type of a connection, the connection provider a `GetType` method:

```go
for _, conn := range mpSock.UnderlaySocket.GetConnections() {
    //stay at incoming connection.
    if conn.GetType() == packets.ConnectionTypes.Incoming {
        // Do Reading here
    }
}

```

### Metrics
The MPPeerSock collects metrics for each connection. In addition to the information SCION already provides (number of hops, latency), the incoming and outgoing bandwidth of connections is measured over their lifetime. The metrics are represented by the following interface:

```go
type PathMetrics struct {
	ReadBytes        int64
	ReadPackets      int64
	WrittenBytes     int64
	WrittenPackets   int64
	ReadBandwidth    []int64 // bytes per MetricsInterval
	WrittenBandwidth []int64 // bytes per MetricsInterval
    // ... further internal fields
}
```

Depending on the `MetricsInterval` field of each `MPPeerSock` instance (by default 1 second), the arrays `ReadBandwidth` and `WrittenBandwidth` are filled. To fetch metrics from the socket, please use the `conn.GetMetrics` method:

```go
for _, conn := range t.Conns {
    m := conn.GetMetrics()
}
```
