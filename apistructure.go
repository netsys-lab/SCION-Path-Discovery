// The package main contains some implementation notes, while the rest of the project is split into sub-packages
// package main

// This file only contains comments, it gives a description of how the
// presented interfaces interact with eachother

// High-level overview:
//
// Client program (Bittorrent)
// |-> Fetches peers from tracker, creates MPPeerSock instances for each peers
//     These instances perform path querying automatically and pass all paths
//     to the PathEnumerator
//     |-> The PathEnumerator combines possible paths and peers and inserts the
//         created PathlevelPeers into the QualityDatabase
//         |-> MPPeerSock creates a PacketScheduler (containing Generator/Handler) and inserts
//             the first path-selection result (Pathlevel Peers with qualities) into the PacketScheduler
//             |-> The MPPeerSock is set to state READY
// |-> Client starts writing data (e.g. Piece Request from Bittorrent) calling MPPeerSock.Write with only payload
//     |-> MPPeerSock forwards the data to the PacketScheduler, which decides over which of the
//         previously selected paths the data should be transferred.
//         |-> PacketScheduler calls PacketGen.Generate creating a SCION packet with the respective path
//             Metrics are increased automatically
//             |-> PacketScheduler writes the SCION packet to the underlying network socket
// |-> Client starts reading data (e.g. reading of piece data) calling MPPeerSock.Read
//     |-> The underlying Socket reads a packet and forwards it to the PacketHandler inside of the PacketScheduler.
//         Packethandler extracts the payload of the SCION packet. Metrics are increased automatically.
//         |-> Packetscheduler passes the payload up to the client who called MPPeerSock.Read
// |-> Asynchronous background routine: Fetches metrics from PacketScheduler and inserts them into
//     Path Quality Database.
//     |-> Path Selection routine is called which gets all collected information to decide which PathlevelPeers
//         Should be passed to the PacketScheduler
