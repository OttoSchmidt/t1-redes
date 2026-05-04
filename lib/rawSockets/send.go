package rawsockets

import (
	"errors"
	"fmt"
	debug "pacman-redes/lib/debug"
	"syscall"
)

/*
Envia a mensagem pelo socket especificado. Não possui timeout, retransmissão ou verificação de resposta.
*/
func sendPacket(sock int, packet *Message) error {
	frame := packet.ToBytes()

	// salvar ultima mensagem enviada
	ServerState.lastSentMessage = *packet

	_, err := syscall.Write(sock, frame)

	fmt.Printf("[MSG] enviado  => %s\n", packet.String())

	return err
}

/*
Reenvia os bytes da última mensagem enviada pelo socket especificado. Útil para reenviar ACK/NACK
após receber um pacote duplicado (retransmissão do remetente).
*/
func ResendLastSent(sock int) error {
	if ServerState.lastSentMessage.EqualsTo(Message{}) {
		return fmt.Errorf("nenhuma mensagem enviada anteriormente")
	}

	frame := ServerState.lastSentMessage.ToBytes()
	_, err := syscall.Write(sock, frame)
	return err
}

/*
Envia a mensagem pelo socket especificado. Se o tipo da mensagem for ACK, NACK ou Error, é enviada sem aguardar resposta.
Para outros tipos, implementa um mecanismo de retransmissão com timeout e limite de tentativas, aguardando alguma resposta.
*/
func SendMessage(sock int, packet *Message) error {
	if (packet.PacketType == Ack || packet.PacketType == Nack || packet.PacketType == Error) {
		err := sendPacket(sock, packet)
		if err != nil {
			return fmt.Errorf("falha ao enviar mensagem: %w", err)
		}

		return nil
	} else {
		timeoutMillis := initialTimeoutMillis

		for attempt := 1; attempt <= MaxAttempts; attempt++ {
			err := sendPacket(sock, packet)
			if err != nil {
				return fmt.Errorf("falha ao enviar mensagem: %w", err)
			}

			debug.PrintLog("Tentativa %d/%d: enviado %d bytes (seq=%d); aguardando ACK por %dms\n", attempt, MaxAttempts, packet.Size(), packet.Sequence, timeoutMillis)

			msg, err := ReceivePacketTWithTimeout(sock, timeoutMillis, Ack)
			
			switch {
			case errors.Is(err, ErrUnexpectedPacketType):
				if msg.PacketType == Nack {
					// reenviar a mensagem, resetando o numero de tentativas
					timeoutMillis = initialTimeoutMillis
					attempt = 0
					continue
				}
			case errors.Is(err, ErrTimeout):
				fmt.Printf("\t- sem resposta dentro de %4dms; reenviando...\n", timeoutMillis)
				timeoutMillis = min(timeoutMillis*2, maxTimeoutMillis)
				continue
			case err == nil:
				debug.PrintLog("\t- ack recebido\n")
				return nil
			default:
				return fmt.Errorf("erro ao aguardar ACK: %w", err)
			}
		}

		return fmt.Errorf("falha ao obter ACK: limite de tentativas atingido")
	}
}

/*
Envia o conteúdo pelo socket especificado. Pode dividir o conteudo em varias mensagens.
Apos a transmissao, envia uma outra mensagem de tipo End para sinalizar o fim
*/
func SendContent(sock int, content []byte, pktType PacketT) error {
	// separar o conteudo em partes, se necessario
	var dataToSend [][]byte
	if len(content) >= maxPacketSize {
		for i := 0; i < len(content); i += maxPacketSize {
			upperBound := i+maxPacketSize
			if upperBound > len(content) {
				upperBound = len(content)
			}
			dataToSend = append(dataToSend, content[i:upperBound])
		}
	} else {
		dataToSend = append(dataToSend, content)
	}

	// enviar as mensagens
	for i := 0; i < len(dataToSend); i++ {
		msg := CreateMessage(dataToSend[i], pktType)
		err := SendMessage(sock, &msg)
		if err != nil {
			return fmt.Errorf("erro ao enviar conteudo: %w\n", err)
		}
	}

	// enviar mensagem de fim
	endMsg := CreateMessage(nil, End)
	err := SendMessage(sock, &endMsg)
	if err != nil {
		return fmt.Errorf("erro ao enviar mensagem de fim: %w\n", err)
	}

	return nil
}