package main

import (
	"fmt"
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
		msg := rawsockets.CreateMessage(fmt.Sprintf("PACMAN-TEST-PACKET-%d", i), rawsockets.PacketTypeData)
		err = rawsockets.SendMessage(sock, msg)
		if err != nil {
			panic(err)
		}
	}
}
