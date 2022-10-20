# SCION Multipath Library - Developer Guide

This documentation is targeted for developers that want to write SCION application based on the SCION Multipath Library (smp). 

**Further resources**:
- [Documetation of concept](https://github.com/netsys-lab/scion-path-discovery/blob/main/doc/path-selection.org)
- [Documentation of implementation](https://github.com/netsys-lab/scion-path-discovery/blob/main/doc/library.md)
- [Go doc](https://pkg.go.dev/github.com/netsys-lab/scion-path-discovery)

## Complete Examples
Before diving into explaining particular functionalities of the library, for an easy start, please refer to the existing examples that already implement smp:
- [MPPingPong](https://github.com/netsys-lab/scion-path-discovery/blob/main/examples/mppingpong/main.go): A program that sends and receives ping and pong messages to other MPPingPong instances using up to `n` paths
- [Simple](../examples/simple/main.go): A simple program sending packets between two PanSockets
- [BitTorrent over SCION](https://github.com/netsys-lab/bittorrent-over-scion): Our BitTorrent over SCION implementation that uses smp for multipath bandwidth aggregation
- [Disjoint](../examples/disjoint/main.go): A benchmark tool sending and receiving as many packets as possible between two PanSockets using our partially disjoint path-selection

## Usage

### Creating Listening PanSockets
To create a PanSocket, initialize it via `smp.NewPanSock` passing the local SCION address as a string to it. The second argument is the remote addr, which should be omitted for sockets that wait for incoming connections. Each instantiated socket must call `Listen`. Afterwards, socket that wait for incoming connections, call `WaitForPeerConnect`. Passing `nil` to this call means, that the peer that connects to this socket performs the path selection. Each PanSock is designed to be connected to a single remote PanSock, creating a 1:1 connection that allows using a variable number of paths.

```go
mpSock := smp.NewPanSock(*localAddr, nil, nil)
err = mpSock.Listen()
if err != nil {
    log.Fatal("Failed to listen PanSock", err)
}

// Wait for incoming connections
remoteAddr, err = mpSock.WaitForPeerConnect(nil)

if err != nil {
    return nil, err
}
```

The third argument provides options to configure the socket via `PanSocketOptions`. This argument may be omitted to apply to default configuration. Default Transport is "QUIC", MultiportMode is disabled by default.

```go
type PanSocketOptions struct {
    // Defines which underlying protocol should be used. 
	Transport                   string // "QUIC"
}

```
### Connecting to PanSockets
To connect to a remote PanSocket, a new instance needs to be created and `Listen` must be called. Furthermore, the peer address needs to be passed to the the constructor, in the form of a `snet.UDPAddr`. After calling Listen, using the `Connect` method, a multipath connection to a waiting remote socket will be established. As first argument, an implementation of the `CustomPathSelection` interface needs to be passed to enable path selection. 

```go
peerAddr, err := snet.ParseUDPAddr(*remoteAddr)
if err != nil {
    log.Fatalf("Failed to parse remote addr %s, err: %v", *remoteAddr, err)
}
mpSock := smp.NewPanSock(*localAddr, peerAddr, nil)
err = mpSock.Listen()
paths, _ := mpSock.GetAvailablePaths()
pathset := pathselection.WrapPathset(paths)
pathset.Address = *peerAddr
err = mpSock.Connect(&pathset, nil)
```

The second argument to Connect are potential `ConnectOptions`. These should be omitted to enable default configuration. However, for fine tuning, the following options are possible:

```go
type ConnectOptions struct {
	SendAddrPacket          bool // Send a packet to the waiting socket containing local information
	DontWaitForIncoming     bool // Force Listening Socket not to wait for incoming connections
	NoPeriodicPathSelection bool // Disable automatic periodic path selection
}
```

### Handle multiple incoming PanSockets
TODO: Just connect back

### Using the instantiated Connections
After connecting to a peer using the `Connect` method, a slice of connections can be fetched via `sock.UnderlaySocket.GetConnections`, where each connection uses one of the selected paths internally. The library provides the channel `OnConnectionsChange` that returns new connections each time an internal ticker starts performing new pathselection. Each connection has a `GetId` method which returns its unique identifier (and also the one of its underlying path), so applications can check if the returned connections changed. The library does not dial again over already used paths. An example use of those methods is shown below:

```go
for _, conn := range mpSock.UnderlaySocket.GetConnections() {
 // ...
}

```


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
In [library.md](https://github.com/netsys-lab/scion-path-discovery/blob/main/doc/library.md#connection-types) we explained the two transport types: SCION and QUIC and how their connections behave. For SCION, there is always one incoming connection, and potential `n` outgoing connections. For two connected PanSockets, the number of outoging connections, and consequently, the number of used paths for outgoing traffic, may differ (as can be seen in the MPPingPong example). Using QUIC as transport type, the connections are always bidirectional, meaning both connected PanSocks have the same number of paths in use. We implement this using a `ConnectionTypes` enum:

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
The PanSock collects metrics for each connection. In addition to the information SCION already provides (number of hops, latency), the incoming and outgoing bandwidth of connections is measured over their lifetime. The metrics are represented by the following interface:

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

Depending on the `MetricsInterval` field of each `PanSock` instance (by default 1 second), the arrays `ReadBandwidth` and `WrittenBandwidth` are filled. To fetch metrics from the socket, please use the `conn.GetMetrics` method:

```go
for _, conn := range t.Conns {
    m := conn.GetMetrics()
}
```
