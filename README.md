# SCION Path Discovery

The goal of the SCION-Pathdiscovery project is to leverage the path-awareness features of the SCION secure Internet architecture in order to discover a fast and reliable content lookup and delivery. The challenge is to search suitable paths considering the respective performance requirements, and combine different candidates into a multi-path connection. To this end, we aim to design and implement optimal path selection and data placement strategies for a decentralized content storage and delivery system.

The primary result is a library that can be used by developers to discover a set of optimised paths across a SCION network. This will first be tested with multiple, high performance paths to a single peer, and subsequently be expanded to larger topologies. As a further demonstration and stress test of these capabilities, the project will implement a real-world use case: BitTorrent over SCION. This should give a good impression of the ability to improve lookup speed and optimise download performance by evaluating and aggregating multiple paths to short lived peers typical to the use pattern of Bittorrent.

Peers are to be identified by a combination of their address and the path used to transfer packets to them, called path-level peers. This approach allows BitTorrent’s sophisticated file sharing mechanisms to run on path level, instead of implementing a dedicated multipath connection to each peer.

## Task 1. Optimal Path Selection for Efficient Multipath Usage
This task consists of creating a detailed concept as well as planning the implementation of a high performance library providing optimal path selection for efficient multipath usage over SCION capable of dealing with the high requirements of BitTorrent

### Milestone(s)
•	Concept of an optimal path selection for efficient multipath usage
•	Architecture design including components, algorithms and design of the implementation


