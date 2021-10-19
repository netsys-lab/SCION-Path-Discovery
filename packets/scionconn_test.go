package packets

import (
	"testing"

	"github.com/netsec-ethz/scion-apps/pkg/appnet"
)

func Test_SCIONConn_Listen(t *testing.T) {
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

}
