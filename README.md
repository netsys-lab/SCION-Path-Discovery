# SCION Path Discovery

The goal of the SCION-Pathdiscovery project is to leverage the path-awareness features of the SCION secure Internet architecture in order to discover a fast and reliable content lookup and delivery. The challenge is to search suitable paths considering the respective performance requirements, and combine different candidates into a multi-path connection. To this end, we aim to design and implement optimal path selection and data placement strategies for a decentralized content storage and delivery system.

The primary result is a library that can be used by developers to discover a set of optimised paths across a SCION network. This will first be tested with multiple, high performance paths to a single peer, and subsequently be expanded to larger topologies. As a further demonstration and stress test of these capabilities, the project will implement a real-world use case: BitTorrent over SCION. This should give a good impression of the ability to improve lookup speed and optimise download performance by evaluating and aggregating multiple paths to short lived peers typical to the use pattern of Bittorrent.

Peers are to be identified by a combination of their address and the path used to transfer packets to them, called path-level peers. This approach allows BitTorrent’s sophisticated file sharing mechanisms to run on path level, instead of implementing a dedicated multipath connection to each peer.

## Task 1. Optimal Path Selection for Efficient Multipath Usage
This task consists of creating a [detailed concept](path-selection/path-selection.org) as well as [planning the implementation](https://godocs.io/github.com/netsys-lab/scion-multipath-lib) [![GoDoc](https://godoc.org/github.com/netsys-lab/scion-multipath-lib?status.svg)](https://godocs.io/github.com/netsys-lab/scion-multipath-lib) of a [high performance library](https://github.com/netsys-lab/scion-multipath-lib) providing optimal path selection for efficient multipath usage over SCION capable of dealing with the high requirements of BitTorrent

### Milestone(s)
-	[Concept](path-selection/path-selection.org) of an optimal path selection for efficient multipath usage
-	[Architecture design](https://godocs.io/github.com/netsys-lab/scion-multipath-lib) including components, algorithms and design of the implementation

## Task 2. Implement Efficient Multipath over SCION
This task contains the implementation of efficient multipath for BitTorrent over SCION in a portable software library.

### Milestone(s)
-	Working implementation

## Task 3. Demonstrate library with BitTorrent over SCION
This task includes the implementation of a demonstrator for the efficient multipath library developed in task 1 and 2 - BitTorrent over SCION - as well as adding a test suite for the library.

### Milestone(s)
-	Working BitTorrent over SCION implementation
-	Continuous integration based testing

## Task 4. Evaluation of Efficient Multipath in SCION library
This task consists of performance measurements for different multipath scenarios, and processing the recommendations made from the security quickscan by Radically Open Security. Additionally, proper documentation is added so that third party developers can start using the library.

### Milestone(s)
-	Evaluation of test results and measurements and process outcome of security quickscan
-	Release 1.0 version of the library
-	Developer documentation

This work is funded through the NGI0 Discovery Fund, a fund established by NLnet with financial support from the European Commission's Next Generation Internet programme, under the aegis of DG Communications Networks, Content and Technology under grant agreement No 825322 https://nlnet.nl/project/SCION-Swarm/
