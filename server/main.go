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

	sock, err := rawsockets.CreateRawSocket(ifaceName)
	defer syscall.Close(sock)
	if err != nil {
		panic(err)
	}

	buf := make([]byte, 65535)

	for {
		n, _, err := syscall.Recvfrom(sock, buf, 0)
		if err != nil {
			panic(err)
		}

		content, id, packetType, crc, err := rawsockets.ReadMessage(buf[:n], n)
		if err != nil {
			fmt.Printf("Erro ao ler mensagem: %v\n", err)
			continue
		}

		fmt.Printf("Pacote capturado: %d bytes\n", n)
		fmt.Printf("ID: %d, Tipo: %d, CRC: %d\n", id, packetType, crc)
		fmt.Printf("Conteúdo: %s\n\n", content)
	}
}