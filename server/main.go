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
		n, addr, err := syscall.Recvfrom(sock, buf, 0)
		if err != nil {
			panic(err)
		}

		if llAddr, ok := addr.(*syscall.SockaddrLinklayer); 
			ok && llAddr.Pkttype == syscall.PACKET_OUTGOING {
			// Ignora pacotes enviados. eles aparecem no loopback, 
			// mas não em interfaces físicas.
			continue 
		}

		content, err := rawsockets.ReadMessage(buf, n)
		if err != nil {
			debug.PrintLog("Erro ao ler mensagem: %v\n", err)
			continue
		}

		fmt.Printf("Conteúdo: %s\n\n", content)
	}
}