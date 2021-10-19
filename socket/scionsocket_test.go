package socket

import (
	"testing"
)

func Test_SCIONSocket_Listen(t *testing.T) {
	t.Run("SCIONConn Listen", func(t *testing.T) {
		sock := NewSCIONSocket("1-ff00:0:110,[127.0.0.12]:41000")
		err := sock.Listen()
		if err != nil {
			t.Error(err)
		}
		sock.CloseAll()
	})

}
