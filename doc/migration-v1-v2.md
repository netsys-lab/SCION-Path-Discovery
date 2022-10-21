# Migration Guide from v1.x to v2.0
While using the scion-path-discovery library in some of our recent SCION projects, we found some initial design decisions that we want to improve with a version 2.0. This document explains these changes, categorized into breaking API changes, and internal changes that may affect the behavior of the library. We updated all examples to version 2.0, so for a quick look how the new API looks like, please refer to the examples folder.

## Breaking API Changes
The following changes are important for applications that use scion-path-discovery version 1.x and should be updated to version 2.x, explaining which core API's changed.

### Rename MPSock to PanSock
The first breaking API change is the renaming of the MPSocket struct to PanSocket. With respect to the term "multipath Socket", this library does not provide a component that does automatic distribution of data over multiple paths. It's core focus is on establishing connections over multiple paths, creating metrics and assisting developers in implementing path-selection. However, applications need to configure a number of connections that they can handle, and read/write data over these connections. Each connection has always a single path applied to each packet, whereas these paths may of course change over time. Also the number of connections may change. Consequently, we decided to rename the MPSocket to PanSocket, to clarify that it is a path-aware socket, but not a dedicated multipath socket.

### Remove Ticker-Based Path-Selection and the respective Events
We observe an implication of the initial approach of performing path-selection from a ticker (periodically after a configurable amount of time passed) inside the socket: Applications need to react to the ticker events, which makes the path-selection process always reactive. If applications want to pro-actively do path-selection, they can force the socket to perform this action, but need to be aware of the periodical path-selection task. This leads to confusion which actions are triggering path-selection, redundant path-selection decisions and  in worst case to oscillation. Consequently, we decided to remove this path-selection approach. 

In version 2.x, sockets provide the following ways to let applications perform path-selection:
1) The `PanSocket.Connect` method expects a pathset as first parameter, defining the number of connections that are established and their respective paths
2) Each connection has a `GetPath()` and `SetPath(path)` method, allowing to freely change paths for each connection
3) If the number of connections needs to change, sockets can simply call `Disconnect()` and `Connect()` again.

### Replace CustomPathSelection Interface with dedicated Structs Wrapping the PanSocket
As explained before, the periodical path-selection has some implications. Furthermore, we observe that passing a struct implementing the original `CustomPathSelection` interface to the socket forces applications to make all path-selection related state accessible from this struct. In our BitTorrent case, this does not perform well, since the path-selection may be performed depending on the information of multiple connected sockets.

Consequently, we decided to remove the `CustomPathSelection` interface and replace it with the path-selection approaches shown above. However, we provide predefined path-selection approaches, i.e. our improved disjoint path-selection, as structs that can be used in combination with the PanSocket to implement path-selection in a straight forward way.


## Internal Changes
The following points are internal changes, that may affect the behavior of the library, but are not visible to application developers.

### Upgrade to scion-apps/pan library
One of the core contributions of version 2.0 is the internal usage of the [scion-apps/pkg/pan](https://github.com/netsec-ethz/scion-apps/tree/master/pkg/pan) library. It provides a stable and well designed interface to implement SCION applications, which we use to provide our path-aware features on top of. In version 1.x, connection dialing/listening, SCION address parsing and many other things were done manually, with not up-to-date versions of external libraries. By moving to pan, we could update these libraries and avoid boilerplate code.

### Bidirectional Handshake to establish Connections
In version 1.x, we sometimes observed that especially SCION/QUIC connections were not established properly. We found out, that quic-go opens streams over the network not when calling `OpenStream`, but when the first data is sent over the stream. This lead to different states in the waiting and connecting socket, and in worst case to some connections that were not usable. In version 2.0, we introduce an explicit handshake over each connection, where the connecting socket sends a handshake packet over each connection and the waiting socket responds with a handshake packet over each connection, to fix exactly this problem.

### Implicit Usage of one Network Socket per Connection
In version 1.x, we introduced the `MultiportMode` flag, which makes each connection have it's own pair of local/remote network sockets. The initial idea behind this was to improve performance, since e.g. in Linux, each socket is bound to a particular CPU and therefor limited in performance. However, we ended in duplicate logic for single/multiport mode support, which was hard to maintain. Consequently, we make the multiport mode default in version 2.0. This has also the nice side effect of easier debugging, because each connection can be identified by its local/remote port numbers.

### Remove of MarkClosed and ConnectionIds
In the original design of scion-path-discovery, connection were marked as "closed" by the socket in case they are not required anymore, so that applications could react to it. However, we observe that returning a specific error while reading/writing on the connection improves the way of reacting to connection closes, because they can be handled immediately in the work loop of the connections. Furthermore, it removes additional, complex logic from the connections.

### Per-Path Metrics 
Instead of per-connection metrics (which were required to be processed to assign the proper path to these metrics), which we defined in version 1.x, we now use per-path metrics directly in the connections. This means each connection is not feeding its own metrics object, but depending on their current path, feed the metrics of this respective path. Furthermore, the PanSocket itself provides functionalities to aggregate metrics, e.g. to find out how specific combinations of paths perform.
