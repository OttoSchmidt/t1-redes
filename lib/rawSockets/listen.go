package rawsockets

import (
	"errors"
	"fmt"
	"syscall"
	"time"

	debug "pacman-redes/lib/debug"
)

/*
Recebe um pacote do socket, aguardando indefinidamente
*/
func ReceivePacket(sock int, buf []byte) (Message, error) {
	n, addr, err := syscall.Recvfrom(sock, buf, 0)
	if err != nil {
		return Message{}, fmt.Errorf("falha ao receber pacote: %w", err)
	}

	if llAddr, ok := addr.(*syscall.SockaddrLinklayer); ok && llAddr.Pkttype == syscall.PACKET_OUTGOING {
		// Ignora pacotes enviados. eles aparecem no loopback,
		// mas não em interfaces físicas.
		return Message{}, ErrIgnoredPacket
	}

	msg, err := ReadMessage(buf, n)
	if err != nil {
		if err != ErrInvalidStartMarker {
			debug.PrintLog("Erro ao ler mensagem: %v\n", err)
		}
			
		return Message{}, err
	}

	WriteMessageLog(fmt.Sprintf("[MSG] recebido => %s\n", msg.String()))

	return msg, nil
}

/*
Recebe um pacote do socket, aguardando um tempo máximo especificado (timeout)
*/
func ReceivePacketWithTimeout(sock int, timeoutMillis int) (Message, error) {
	if timeoutMillis <= 0 {
		return Message{}, fmt.Errorf("timeout invalido: %d", timeoutMillis)
	}

	deadline := time.Now().Add(time.Duration(timeoutMillis) * time.Millisecond)
	buf := make([]byte, 256)

	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return Message{}, ErrTimeout
		}

		if remaining > 150*time.Millisecond {
			remaining = 150 * time.Millisecond
		}

		tv := syscall.NsecToTimeval(remaining.Nanoseconds())
		if err := syscall.SetsockoptTimeval(sock, syscall.SOL_SOCKET, syscall.SO_RCVTIMEO, &tv); err != nil {
			return Message{}, fmt.Errorf("falha ao configurar timeout do socket: %w", err)
		}

		msg, err := ReceivePacket(sock, buf)
		if err != nil {
			switch {
			case errors.Is(err, ErrIgnoredPacket),
				errors.Is(err, ErrInvalidStartMarker),
				errors.Is(err, ErrDuplicatePacket),
				errors.Is(err, syscall.EAGAIN),
				errors.Is(err, syscall.EWOULDBLOCK),
				errors.Is(err, syscall.EINTR):
				continue
			default:
				return Message{}, err
			}
		}

		return msg, nil
	}
}

/*
Recebe um pacote do socket, aguardando um tempo máximo especificado (timeout) e filtrando por um tipo especifico.
*/
func ReceivePacketTWithTimeout(sock int, timeoutMillis int, expectedType PacketT) (Message, error) {
	deadline := time.Now().Add(time.Duration(timeoutMillis) * time.Millisecond)

	remaining := time.Until(deadline)
	if remaining <= 0 {
		return Message{}, ErrTimeout
	}

	msg, err := ReceivePacketWithTimeout(sock, int(remaining/time.Millisecond))
	switch {
	case errors.Is(err, ErrTimeout):
		return Message{}, ErrTimeout
	case err != nil:
		return Message{}, err
	}

	if msg.PacketType != expectedType {
		return msg, ErrUnexpectedPacketType
	}

	return msg, nil
	
}

func ReceiveContent(sock int, buf []byte) ([]byte, error) {
	messageCompleted := false
	content := make([]byte, 0)

	for !messageCompleted {
		msg, err := ReceivePacket(sock, buf)
		if err != nil {
			if errors.Is(err, ErrDuplicatePacket) {
				// pacote duplicado. se o ultimo pacote enviado foi ACK, enviar novamente e ignorar mensagem atual
				if ServerState.lastSentMessage.PacketType == Ack {
					if resendErr := ResendLastSent(sock); resendErr != nil {
						return nil, fmt.Errorf("erro ao reenviar ultimo pacote (ack): %v\n", resendErr)
					}
					return nil, nil
				}
			} else if errors.Is(err, ErrInvalidCRC) {
				// enviar nack e esperar pela mensagem correta
				if resendErr := ResendLastSent(sock); resendErr != nil {
					return nil, fmt.Errorf("erro ao enviar nack: %v\n", resendErr)
				}
				continue
			} else if errors.Is(err, ErrIgnoredPacket) || errors.Is(err, ErrInvalidStartMarker) {
				continue
			} else {
				return nil, err
			}
		}

		// enviar ack
		if msg.PacketType != Ack && msg.PacketType != Nack {
			replyMsg := CreateMessage(nil, Ack)
			if err = SendMessage(sock, &replyMsg); err != nil {
				debug.PrintLog("erro ao enviar ack: %v\n", err)
			}
		}

		switch msg.PacketType {
		case Ack, Nack:
			return nil, nil
		case Data:
			content = append(content, msg.Content...)
		case TxtFile, JpgFile, Mp4File:
			id, tam, err := ParseFileHeader(msg.Content)
			if err != nil {
				return nil, fmt.Errorf("erro ao interpretar cabecalho de arquivo: %v\n", err)
			}

			// ler os pacotes de dado do arquivo
			file, err := ReceiveFile(sock, id, tam, msg.PacketType)
			if err != nil {
				return nil, fmt.Errorf("erro ao receber arquivo: %v\n", err)
			}

			fmt.Printf("arquivo recebido e salvo em: %s\n", file)

			// abrir arquivo com handler padrao do sistema
			if err := OpenDefaultFileHandler(file); err != nil {
				fmt.Printf("erro ao abrir arquivo com handler padrao: %v\n", err)
			}

			messageCompleted = true
		case End:
			messageCompleted = true
		default:
			return nil, fmt.Errorf("tipo de mensagem desconhecido (%d)\n", msg.PacketType)
		}

	}

	return content, nil
}