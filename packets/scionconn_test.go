package packets

import (
	"testing"
	"time"

	"github.com/netsys-lab/scion-path-discovery/sutils"
)

func Test_SCIONConn(t *testing.T) {
	t.Run("SCIONConn Listen", func(t *testing.T) {
		conn := SCIONTransportConstructor()
		addr, err := sutils.ResolveUDPAddr("19-ffaa:1:cf1,[127.0.0.1]:41000")
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
		addr, err := sutils.ResolveUDPAddr("19-ffaa:1:cf1,[127.0.0.1]:42000")
		if err != nil {
			t.Error(err)
		}
		laddr, err := sutils.ResolveUDPAddr("19-ffaa:1:cf1,[127.0.0.1]:43000")
		if err != nil {
			t.Error(err)
		}
		conn.SetLocal(*laddr)
		err = sutils.SetDefaultPath(addr)
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
		addr, err := sutils.ResolveUDPAddr("19-ffaa:1:cf1,[127.0.0.1]:44000")
		if err != nil {
			t.Error(err)
		}
		laddr, err := sutils.ResolveUDPAddr("19-ffaa:1:cf1,[127.0.0.1]:45000")
		if err != nil {
			t.Error(err)
		}
		conn.SetLocal(*laddr)
		err = sutils.SetDefaultPath(addr)
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
