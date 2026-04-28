package main

import (
	"errors"
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
			if errors.Is(err, rawsockets.ErrDuplicatePacket) {
				// pacote duplicado (retransmissão após ACK perdido): reenviar último ACK/NACK
				if resendErr := rawsockets.ResendLastSent(sock); resendErr != nil {
					debug.PrintLog("Erro ao reenviar resposta para pacote duplicado: %v\n", resendErr)
				}
			} else {
				debug.PrintLog("Erro ao receber pacote: %v\n", err)
			}
			continue
		}

		fmt.Printf("Conteúdo: %s\n\n", msg.Content)

		switch msg.PacketType {
		case rawsockets.Ack, rawsockets.Nack:
			continue
		case rawsockets.Data:
			msg := rawsockets.CreateMessage(nil, rawsockets.Ack)
			if err := rawsockets.SendMessage(sock, &msg); err != nil {
				debug.PrintLog("Erro ao enviar ACK: %v\n", err)
			}
		case rawsockets.TxtFile, rawsockets.JpgFile, rawsockets.Mp4File:
			// enviar ack/nack
			reply := rawsockets.CreateMessage(nil, rawsockets.Ack)
			if err := rawsockets.SendMessage(sock, &reply); err != nil {
				debug.PrintLog("Erro ao enviar ACK: %v\n", err)
			}

			// extrair id, tamanho e tipo do arquivo do conteúdo do pacote
			id, tam, err := rawsockets.ParseFileHeader(msg.Content)
			if err != nil {
				debug.PrintLog("Erro ao parsear cabecalho do arquivo: %v\n", err)
				continue
			}

			// receber arquivo e salvar em disco
			file, err := rawsockets.ReceiveFile(sock, id, tam, msg.PacketType)
			if err != nil {
				debug.PrintLog("Erro ao receber arquivo: %v\n", err)
				continue
			}
			fmt.Printf("Arquivo recebido e salvo em: %s\n", file)

			// abrir arquivo com handler padrao do sistema
			if err := rawsockets.OpenDefaultFileHandler(file); err != nil {
				debug.PrintLog("Erro ao abrir arquivo com handler padrao: %v\n", err)
			}
		case rawsockets.End:
			msg := rawsockets.CreateMessage(nil, rawsockets.Ack)
			if err := rawsockets.SendMessage(sock, &msg); err != nil {
				debug.PrintLog("Erro ao enviar ACK: %v\n", err)
			}
		default:
			msg := rawsockets.CreateMessage(nil, rawsockets.Nack)
			if err := rawsockets.SendMessage(sock, &msg); err != nil {
				debug.PrintLog("Erro ao enviar NACK: %v\n", err)
			}
		}
	}
}
