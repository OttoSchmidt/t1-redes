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
	if err != nil {
		panic(err)
	}
	defer syscall.Close(sock)

	// enviar pacote de teste para o servidor
	n, err := rawsockets.SendMessage(sock, "PACMAN-TEST-PACKET", 0)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Pacote de teste enviado: %d bytes na interface %s\n", n, ifaceName)
}