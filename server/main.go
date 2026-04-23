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

	sock, err := rawsockets.CreateRawSocket(ifaceName)
	defer syscall.Close(sock)
	if err != nil {
		panic(err)
	}

	buf := make([]byte, 36)

	fmt.Println("Servidor iniciado. Esperando mensagens...")
	for {
		_, addr, err := syscall.Recvfrom(sock, buf, 0)
		if err != nil {
			panic(err)
		}

		if llAddr, ok := addr.(*syscall.SockaddrLinklayer); ok && llAddr.Pkttype == syscall.PACKET_OUTGOING {
			// Ignora pacotes enviados. eles aparecem no loopback,
			// mas não em interfaces físicas.
			continue
		}

		msg, err := rawsockets.ReadPacket(buf)
		if err != nil {
			if err != rawsockets.ErrInvalidStartMarker {
				debug.PrintLog("Erro ao ler mensagem: %v\n", err)
			}
				
			continue
		}

		fmt.Printf("Conteúdo: %s\n\n", msg.Content)

		switch msg.PacketType {
		case rawsockets.PacketTypeAck, rawsockets.PacketTypeNack:
			continue
		case rawsockets.PacketTypeData:
			reply := rawsockets.CreateMessage("", rawsockets.PacketTypeAck)
			if _, err := rawsockets.SendMessage(sock, reply); err != nil {
				debug.PrintLog("Erro ao enviar ACK: %v\n", err)
			}
		default:
			reply := rawsockets.CreateMessage("", rawsockets.PacketTypeNack)
			if _, err := rawsockets.SendMessage(sock, reply); err != nil {
				debug.PrintLog("Erro ao enviar NACK: %v\n", err)
			}
		}
	}
}
