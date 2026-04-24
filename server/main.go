package main

import (
	"fmt"
	"os"
	"syscall"

	debug "pacman-redes/lib/debug"
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
	defer syscall.Close(sock)
	if err != nil {
		panic(err)
	}

	buf := make([]byte, 256)

	fmt.Println("Servidor iniciado. Esperando mensagens...")
	for {
		msg, err := rawsockets.ReceivePacket(sock, buf)
		if err != nil {
			debug.PrintLog("Erro ao receber pacote: %v\n", err)
			continue
		}

		fmt.Printf("Conteúdo: %s\n\n", msg.Content)

		switch msg.PacketType {
		case rawsockets.Ack, rawsockets.Nack:
			continue
		case rawsockets.Data:
			reply := rawsockets.CreateMessage("", rawsockets.Ack)
			if err := rawsockets.SendMessage(sock, reply); err != nil {
				debug.PrintLog("Erro ao enviar ACK: %v\n", err)
			}
		default:
			reply := rawsockets.CreateMessage("", rawsockets.Nack)
			if err := rawsockets.SendMessage(sock, reply); err != nil {
				debug.PrintLog("Erro ao enviar NACK: %v\n", err)
			}
		}
	}
}
