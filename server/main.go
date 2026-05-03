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
	defer syscall.Close(sock)
	if err != nil {
		panic(err)
	}

	buf := make([]byte, 256)

	fmt.Println("Servidor iniciado. Esperando mensagens...")
	for {
		content, err := rawsockets.ReceiveContent(sock, buf)
		if err != nil {
			fmt.Printf("erro ao receber conteudo:\n\t- %v\n", err)
		}

		fmt.Printf("conteudo: %s\n", content)
	}
}
