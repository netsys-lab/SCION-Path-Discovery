package packets

/*
// TODO: Implement SCION/QUIC here
type QUICReliableConn struct { // Former: MonitoredConn
	internalConn *quic.Stream
	path         *snet.Path
	peer         string
	state        int // See Connection States
	metrics      PacketMetrics
}

// This simply wraps conn.Read and will later collect metrics
func (qc *QUICReliableConn) Read(b []byte) (int, error) {
	n, err := qc.internalConn.Read(b)
	if err != nil {
		return n, err
	}
	qc.metrics.ReadBytes += int64(n)
	qc.metrics.ReadPackets++
	return n, err
}

// This simply wraps conn.Write and will later collect metrics
func (qc *QUICReliableConn) Write(b []byte) (int, error) {
	n, err := qc.internalConn.Write(b)
	qc.metrics.WrittenBytes += int64(n)
	qc.metrics.WrittenPackets++
	if err != nil {
		return n, err
	}
	return n, err
}

func (qc *QUICReliableConn) WriteStream(b []byte) (int, error) {
	bts := make([]byte, 8)
	binary.BigEndian.PutUint64(bts, uint64(len(b)))
	n, err := qc.Write(bts)
	if err != nil {
		return n, err
	}

	n, err = qc.Write(b)
	return n, err

}

func (qc *QUICReliableConn) ReadStream(b []byte) (int, error) {
	bts := make([]byte, 8)
	n, err := qc.Read(bts)
	if err != nil {
		return n, err
	}
	len := binary.BigEndian.Uint64(bts)
	buf := make([]byte, 9000)
	b = make([]byte, len)
	var i uint64
	io.ReadFull()
	for i < len {
		n, err := qc.Read(buf)
		if err != nil {
			return int(i), err
		}
		copy(b[i:int(i)+n], buf)
		i += uint64(n)
	}

	return int(i), err

}

func (qc *QUICReliableConn) Listen(snet.UDPAddr) {
	l, err := appquic.Listen(
		&net.UDPAddr{},
		&tls.Config{
			Certificates: appquic.GetDummyTLSCerts(),
			NextProtos:   []string{"scion-filetransfer"},
		},
		&quic.Config{
			KeepAlive: true,
		},
	)

}
*/
