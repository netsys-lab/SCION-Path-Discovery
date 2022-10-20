package socket

import (
	"testing"

	lookup "github.com/netsys-lab/scion-path-discovery/pathlookup"
	"github.com/netsys-lab/scion-path-discovery/pathselection"
)

func Test_QUICSocket(t *testing.T) {
	t.Run("QUICSocket Listen", func(t *testing.T) {
		sock := NewQUICSocket("1-ff00:0:110,[127.0.0.12]:31000")
		err := sock.Listen()
		if err != nil {
			t.Error(err)
		}
		sock.CloseAll()
	})

	t.Run("SCIONSocket Listen And Dial", func(t *testing.T) {
		sock := NewQUICSocket("1-ff00:0:110,[127.0.0.12]:21100")
		err := sock.Listen()
		if err != nil {
			t.Error(err)
			return
		}
		defer sock.CloseAll()

		sock2 := NewQUICSocket("1-ff00:0:110,[127.0.0.12]:11100")
		err = sock2.Listen()
		if err != nil {
			t.Error(err)
			return
		}

		go func() {
			paths, err := lookup.PathLookup("1-ff00:0:110,[127.0.0.12]:21100")
			if err != nil {
				t.Error(err)
				return
			}

			if len(paths) == 0 {
				t.Error("No paths found for local AS, something is wrong here...")
				return
			}

			pathQualities := make([]pathselection.PathQuality, 1)
			pathQualities[0] = pathselection.PathQuality{
				Id:       "FirstPath",
				SnetPath: paths[0],
			}

			sock2.DialAll(*sock.localAddr, pathQualities, DialOptions{SendAddrPacket: true})
		}()

		_, err = sock.WaitForDialIn()
		if err != nil {
			t.Error(err)
			return
		}
		sock.CloseAll()
		sock2.CloseAll()
	})

}
