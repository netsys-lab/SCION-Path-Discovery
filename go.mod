module github.com/netsys-lab/scion-path-discovery

go 1.16

require (
	github.com/lucas-clemente/quic-go v0.21.1
	github.com/netsec-ethz/scion-apps v0.3.1-0.20210924130723-be84cbd98c1f
	github.com/netsys-lab/scion-optimized-connection v0.4.2-0.20220107124242-cc4b4825db7f
	github.com/scionproto/scion v0.6.0
	github.com/sirupsen/logrus v1.8.1
)

replace github.com/netsys-lab/scion-optimized-connection => /home/marten/go/src/github.com/netsys-lab/scion-optimized-connection
