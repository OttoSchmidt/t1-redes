package main

import (
	"errors"
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

	sock, err := rawsockets.CreateRawSocket(ifaceName)
	if err != nil {
		panic(err)
	}
	defer syscall.Close(sock)

	const maxAttempts = 4
	timeoutMillis := 1000
	message := "PACMAN-TEST-PACKET"

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		sequence := uint8((attempt - 1) & 0x3F)
		n, err := rawsockets.SendMessageWithSequence(sock, message, sequence, rawsockets.PacketTypeData)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Tentativa %d/%d: enviado %d bytes na interface %s (seq=%d); aguardando ACK por %dms\n", attempt, maxAttempts, n, ifaceName, sequence, timeoutMillis)

		ack, err := rawsockets.ReceivePacketTypeWithTimeout(sock, timeoutMillis, rawsockets.PacketTypeAck)
		if err == nil {
			fmt.Printf("ACK recebido: %s\n", ack.Content)
			return
		}

		if errors.Is(err, rawsockets.ErrTimeout) {
			fmt.Printf("Sem ACK dentro de %dms; reenviando...\n", timeoutMillis)
			timeoutMillis *= 2
			continue
		}

		panic(err)
	}

	panic("falha ao obter ACK: limite de tentativas atingido")
}
