package main

import (
	"fmt"
	"os"
	"syscall"
	"time"

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

	// esperar janela de logs
	time.Sleep(3 * time.Second)

	buf := make([]byte, 256)

	fmt.Println("Esperando mensagens...")
	for {
		content, err := rawsockets.ReceiveContent(sock, buf)
		if err != nil {
			fmt.Printf("erro ao receber conteudo:\n\t- %v\n", err)
		}

		if len(content) > 0 {
			rawsockets.WriteMessageLog(fmt.Sprintf("Conteudo: %s\n", content))
		}
	}
}
