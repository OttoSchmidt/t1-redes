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

	sock, err := rawsockets.CreateRawSocket(ifaceName)
	if err != nil {
		panic(err)
	}
	defer syscall.Close(sock)

	rawsockets.AttemptSendMessage(sock, rawsockets.Message{
		Content:    "PACMAN-TEST-PACKET",
		Sequence:   rawsockets.SequenceNumber,
		PacketType: rawsockets.PacketTypeData,
	})

	panic("falha ao obter ACK: limite de tentativas atingido")
}
