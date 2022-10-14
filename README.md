# SCION Path Discovery
[![Go Reference](https://pkg.go.dev/badge/github.com/netsys-lab/scion-path-discovery.svg)](https://pkg.go.dev/github.com/netsys-lab/scion-path-discovery)
[![License](https://img.shields.io/github/license/netsys-lab/scion-path-discovery.svg?maxAge=2592000)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/netsys-lab/scion-path-discovery)](https://goreportcard.com/report/github.com/netsys-lab/scion-path-discovery)


The goal of the SCION-Pathdiscovery project is to leverage the path-awareness features of the SCION secure Internet architecture in order to discover a fast and reliable content lookup and delivery. The challenge is to search suitable paths considering the respective performance requirements, and combine different candidates into a multi-path connection. To this end, we aim to design and implement optimal path selection and data placement strategies for a decentralized content storage and delivery system.

The primary result is a library that can be used by developers to discover a set of optimised paths across a SCION network. This will first be tested with multiple, high performance paths to a single peer, and subsequently be expanded to larger topologies. As a further demonstration and stress test of these capabilities, the project will implement a real-world use case: BitTorrent over SCION. This should give a good impression of the ability to improve lookup speed and optimise download performance by evaluating and aggregating multiple paths to short lived peers typical to the use pattern of Bittorrent.

Peers are to be identified by a combination of their address and the path used to transfer packets to them, called path-level peers. This approach allows BitTorrent’s sophisticated file sharing mechanisms to run on path level, instead of implementing a dedicated multipath connection to each peer.

## Task 1. Optimal Path Selection for Efficient Multipath Usage
This task consists of creating a detailed [concept](doc/path-selection.org) as well as planning the implementation of a high performance library providing optimal [path selection](https://pkg.go.dev/github.com/netsys-lab/scion-path-discovery/pathselection) for efficient [multipath usage](https://pkg.go.dev/github.com/netsys-lab/scion-path-discovery/api) over [SCION](https://www.scion-architecture.net) capable of dealing with the high requirements of BitTorrent

### Milestone(s)
- [x] [Concept](doc/path-selection.org) of an optimal path selection for efficient multipath usage
- [x] [Architecture design](https://pkg.go.dev/github.com/netsys-lab/scion-path-discovery#section-directories) including components, algorithms and design of the implementation

## Task 2. Implement Efficient Multipath over SCION
This task contains the [implementation](https://github.com/netsys-lab/scion-path-discovery/releases/tag/implementation) of efficient multipath for BitTorrent over SCION in a portable software library.

### Milestone(s)
- [x] Working [implementation](https://github.com/netsys-lab/scion-path-discovery/releases/tag/implementation) 

Further information:
- [x] [Documentation](doc/library.md) of the implemented components and how they work together
- [x] [Example](doc/library.md#example-multipath-pingpong) of how the library can be used to perform multipath communication
- [x] [Nix flake repo for scion-path-discovery](https://github.com/ngi-nix/scion-path-discovery), a packaging of scion-path-discovery for NixOS (work in progress)

## Task 3. Demonstrate library with BitTorrent over SCION
This task includes the [implementation](https://github.com/netsys-lab/bittorrent-over-scion/releases/tag/implementation) of a demonstrator for the efficient multipath library developed in task 1 and 2 - BitTorrent over SCION - as well as adding a test suite for the library.

### Milestone(s)
- [x] [Working BitTorrent over SCION implementation](https://github.com/netsys-lab/bittorrent-over-scion/releases/tag/implementation)
- [x] [Continuous integration based testing](https://github.com/netsys-lab/scion-path-discovery/actions/workflows/test.yml)

## Task 4. Evaluation of Efficient Multipath in SCION library
This task consists of performance measurements for different multipath scenarios, and processing the recommendations made from the security quickscan by Radically Open Security. Additionally, proper documentation is added so that third party developers can start using the library.

### Milestone(s)
- [x] Evaluation of test results and measurements and process outcome of security quickscan
    - Evaluation of test results: Presented in our [Multipath BitTorrent over SCION Paper](doc/bittorrent-over-scion.pdf)
    - Outcome of security quickscan: Fixes in [SCION Pathdiscovery](https://github.com/netsys-lab/scion-path-discovery/pull/27) and [BitTorrent over SCION](https://github.com/netsys-lab/bittorrent-over-scion/pull/5)
- [x] Release 1.0 version of the [SCION Pathdiscovery](https://github.com/netsys-lab/scion-path-discovery/releases/tag/v1.0.0) library and [BitTorrent over SCION](https://github.com/netsys-lab/bittorrent-over-scion/releases/tag/v1.0.0)
- [x] [Developer documentation](doc/developer-doc.md)

### Further Resources
- [x] [Demonstration video](https://drive.google.com/file/d/1zDdmPvLGh1MXgV5Ne1qezudgTcmd7eVq/view?usp=sharing) of BitTorrent over SCION
- [x] [Demo seeder](https://github.com/netsys-lab/bittorrent-over-scion/tree/master/demo) that provides sample torrent running in SCIONLab

## Task 5. A partially disjoint path selection algorithm

We propose to enhance our SCION multipath library with a partially disjoint path selection algorithm, that will significantly increase the application performance by gradually re-distributing application traffic over alternative paths, while monitoring overall throughput.

### Milestone(s)
- [ ] In our SCION Pathdiscovery project funded by NGI Zero Discovery, we have developed robust concepts for disjoint path selection as a means to avoid shared bottlenecks. We achieve bottleneck avoidance in uploading peers through only using completely disjoint paths. However, also partially disjoint paths may lead to performance improvements, which was not considered yet. We thus propose to extend our core path selection strategy as follows: 
  - Peers perform "path exploration" by shifting portions of their traffic to alternative paths while monitoring overall throughput, gradually re-distributing traffic to maximize network utilization. 
  - We expect the proposed extensions to significantly increase the performance of our SCION multipath library by leveraging unused network capacities. 
  - Finally, we anticipate that our refined path selection approach contributes significant additional insights in the field of path-aware networking. 
  - Consequently, we plan to implement and evaluate the proposed path selection algorithm in the next step, in combination with revisiting our multipath library API and performing bug fixes.

This work is funded through the NGI0 Discovery Fund, a fund established by NLnet with financial support from the European Commission's Next Generation Internet programme, under the aegis of DG Communications Networks, Content and Technology under grant agreement No 825322 https://nlnet.nl/project/SCION-Swarm/
