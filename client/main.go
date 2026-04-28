package main

import (
	"os"
	"syscall"

	rawsockets "pacman-redes/lib/rawSockets"
)

func main() {
	// interface padrao: loopback. para usar outra interface,
	// passe o nome como argumento,como: eth0, enp3s0...
	ifaceName := "lo"
	if len(os.Args) > 1 {
		ifaceName = os.Args[1]
	}

	sock, err := rawsockets.CreateSocket(ifaceName)
	if err != nil {
		panic(err)
	}
	defer syscall.Close(sock)

	for i := 0; i < 10; i++ {
		content := "isso eh uma mensagem maior que 31 bytes. o esperado eh que ele divida em varias mensagens."
		err := rawsockets.SendContent(sock, []byte(content), rawsockets.Data)
		if err != nil {
			panic(err)
		}
	}
}
