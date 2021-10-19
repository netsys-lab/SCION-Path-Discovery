package packets

import (
	"testing"
	"time"

	"github.com/netsec-ethz/scion-apps/pkg/appnet"
)

func Test_SCIONConn(t *testing.T) {
	t.Run("SCIONConn Listen", func(t *testing.T) {
		conn := SCIONTransportConstructor()
		addr, err := appnet.ResolveUDPAddr("1-ff00:0:110,[127.0.0.12]:41000")
		if err != nil {
			t.Error(err)
		}
		err = conn.Listen(*addr)
		if err != nil {
			t.Error(err)
		}
		conn.Close()
	})

	t.Run("SCIONConn Dial", func(t *testing.T) {
		conn := SCIONTransportConstructor()
		addr, err := appnet.ResolveUDPAddr("1-ff00:0:110,[127.0.0.12]:42000")
		if err != nil {
			t.Error(err)
		}
		laddr, err := appnet.ResolveUDPAddr("1-ff00:0:110,[127.0.0.12]:43000")
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
		err = conn.Dial(*addr, &p)
		if err != nil {
			t.Error(err)
		}

		defer conn.Close()

	})

	t.Run("SCIONConn Read/Write", func(t *testing.T) {
		conn := SCIONTransportConstructor()
		addr, err := appnet.ResolveUDPAddr("1-ff00:0:110,[127.0.0.12]:44000")
		if err != nil {
			t.Error(err)
		}
		laddr, err := appnet.ResolveUDPAddr("1-ff00:0:110,[127.0.0.12]:45000")
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
		err = conn.Dial(*addr, &p)
		if err != nil {
			t.Error(err)
		}

		listenConn := SCIONTransportConstructor()
		listenConn.Listen(*addr)

		go func() {
			time.Sleep(100 * time.Millisecond)
			conn.Write(make([]byte, 1200))
		}()
		buf := make([]byte, 1200)
		_, err = listenConn.Read(buf)
		if err != nil {
			t.Error(err)
		}
		defer conn.Close()

	})

}
