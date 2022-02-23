package packets

import (
	"testing"
	"time"

	"github.com/netsys-lab/scion-path-discovery/sutils"
)

func Test_QUICConn(t *testing.T) {
	t.Run("QUICConn Listen", func(t *testing.T) {
		conn := QUICConnConstructor()
		udpAddr, err := sutils.ResolveUDPAddr("19-ffaa:1:cf1,[127.0.0.1]:51000")
		if err != nil {
			t.Error(err)
		}
		err = conn.Listen(*udpAddr)
		if err != nil {
			t.Error(err)
		}
		conn.Close()
	})

	t.Run("QUICConn Read/Write", func(t *testing.T) {
		conn := &QUICReliableConn{}
		rudpAddr, err := sutils.ResolveUDPAddr("19-ffaa:1:cf1,[127.0.0.1]:54000")
		if err != nil {
			t.Error(err)
		}
		ludpAddr, err := sutils.ResolveUDPAddr("19-ffaa:1:cf1,[127.0.0.1]:55000")
		if err != nil {
			t.Error(err)
		}
		conn.SetLocal(*ludpAddr)

		err = sutils.SetDefaultPath(rudpAddr)
		if err != nil {
			t.Error(err)
		}
		p, err := rudpAddr.GetPath()
		if err != nil {
			t.Error(err)
		}

		listenConn := &QUICReliableConn{}
		listenConn.Listen(*ludpAddr)

		go func() {
			time.Sleep(100 * time.Millisecond)
			err = conn.Dial(*rudpAddr, &p)
			if err != nil {
				t.Error(err)
				return
			}
			conn.Write(make([]byte, 1200))
		}()
		s, err := listenConn.AcceptStream()
		if err != nil {
			t.Error(err)
		}
		listenConn.SetStream(s)
		buf := make([]byte, 1200)
		_, err = listenConn.Read(buf)
		if err != nil {
			t.Error(err)
		}
		defer conn.Close()

	})

}
