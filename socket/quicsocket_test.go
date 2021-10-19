package socket

import (
	"testing"
)

func Test_QUICSocket(t *testing.T) {
	t.Run("QUICSocket Listen", func(t *testing.T) {
		sock := NewQUICSocket("1-ff00:0:110,[127.0.0.12]:31000", &SockOptions{PathSelectionResponsibility: "server"})
		err := sock.Listen()
		if err != nil {
			t.Error(err)
		}
		sock.CloseAll()
	})

}
