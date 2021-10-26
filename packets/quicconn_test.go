package packets

import (
	"testing"
	"time"

	"github.com/netsec-ethz/scion-apps/pkg/appnet"
)

func Test_QUICConn(t *testing.T) {
	t.Run("QUICConn Listen", func(t *testing.T) {
		conn := QUICConnConstructor()
		addr, err := appnet.ResolveUDPAddr("1-ff00:0:110,[127.0.0.12]:51000")
		if err != nil {
			t.Error(err)
		}
		err = conn.Listen(*addr)
		if err != nil {
			t.Error(err)
		}
		conn.Close()
	})

	t.Run("QUICConn Read/Write", func(t *testing.T) {
		conn := &QUICReliableConn{}
		addr, err := appnet.ResolveUDPAddr("1-ff00:0:110,[127.0.0.12]:54000")
		if err != nil {
			t.Error(err)
		}
		laddr, err := appnet.ResolveUDPAddr("1-ff00:0:110,[127.0.0.12]:55000")
		if err != nil {
			t.Error(err)
		}
		conn.SetLocal(*laddr)
		err = appnet.SetDefaultPath(addr)
		if err != nil {
			t.Error(err)
		}
		p, err := addr.GetPath()
		if err != nil {
			t.Error(err)
		}

		listenConn := &QUICReliableConn{}
		listenConn.Listen(*addr)

		go func() {
			time.Sleep(100 * time.Millisecond)
			err = conn.Dial(*addr, &p)
			if err != nil {
				t.Error(err)
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
