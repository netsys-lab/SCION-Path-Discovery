# Multipath Library - Implementation

## Overview

The following components are compound to implement the multipath library
- **UnderlaySocket**: Provides listen and connect methods that return TransportConnections
- **Connection**: Abstracts underlay connections providing metrics and read/write methods
- **PathQuality**: Stores metrics and meta information about paths
- **PathQualityDB**: Serves as database for PathQuality entries
- **CustomPathSelection**: Provides an interface to implement custom path selection algorithms
- **PanSocket**: Represents a multipath socket to a particular peer

The following figure illustrates, how these components are combined to form a working multipath library, which implements the [path selection concept](https://github.com/netsys-lab/scion-path-discovery/blob/main/doc/path-selection.org#concept).

![pathdiscovery-abstractions (1)](https://user-images.githubusercontent.com/32448709/137099751-ec4233a6-6312-407b-ab94-1139c484029b.jpg)


The application starts the multipath socket by creating a new instance of PanSocket and passes a custom pathselection algorithm optionally. To start communication, the Listen method is called by the application. Afterwards, the PanSocket can either call Connect to another listening socket or call Accept to wait for incoming connections. These connections are instantiated with Connection components. Each Connection runs over a particular path. Data can be read or written using Read and Write methods of one or more connections, concurrently. These connections collect metrics for used paths which are fetched from the PanSocket instance and stored in the PathQualityDB.

## Pathselection
Pathselection can be easily implemented via
1) Passing an initial pathset to the `Connect` method that establishes the connections over these paths.
2) Use `GetPath()` and `SetPath(path)` Methods to change paths on the fly.
3) For changing the number of connections, use `Disconnect` and `Connect`again.

## Transport Types
At the moment, we support two different transports: SCION over plain UDP (SCION/UDP) and SCION over QUIC (SCION/QUIC). Both transports create bidirectional, end-to-end connections. However, SCION/UDP does not provide reliable transport, so the application need to implement retransmission and loss detection. SCION/QUIC has reliability built-in, since it is based on QUIC.


## Using Multiple Paths
After connecting to a peer using the `Connect` method, a slice of connections can be fetched via `sock.UnderlaySocket.GetConnections`, where each connection uses one of the selected paths internally. An example use of those methods is shown below:

```go
for _, conn := range mpSock.UnderlaySocket.GetConnections() {
 // ...
}
```

## Extensibility
We aim the design of this library to be easily extensible for further metrics, Connections or UnderlaySockets. New UnderlaySockets and/or Connections can be added without touching the existing ones and may be added via the socketOptions "Transport" flag. An UnderlaySocket may also be extended to use different Connections, e.g. the [snet](https://github.com/scionproto/scion/tree/master/go/lib/snet) SCION Connection or the [SCION OptimizedConn](https://github.com/netsys-lab/scion-optimized-connection). By introducing the CustomPathSelection interface, applications can easily implement different kinds of pathselection without the need for touching the library, but with helpful utilities to pre-sort paths.

## Example: Multipath PingPong
To test the multipath capabilities of this library, we provide an example, called [multipath pingpong](https://github.com/netsys-lab/scion-path-discovery/blob/main/examples/mppingpong/main.go). This example can be started with a local and a remote SCION address and a number of outgoing connections n. We call one running instance of this example a peer. To see how multipath communication works, two peers need to be started. Each peer sends ping packets over n connections and reads all incoming pong connections, echoing over which paths the pings are sent and over which paths the pongs are received. This example is using SCION/QUIC connections. 

Host 1: `./mppingpong -l "19-ffaa:1:c3f,[141.44.25.148]:32000" -r "19-ffaa:1:cf0,[141.44.25.151]:32000"  -n 2`

Host 2: `./mppingpong -r 19-ffaa:1:c3f,141.44.25.148:32000 -l 19-ffaa:1:cf0,141.44.25.151:32000  -n 3 `
