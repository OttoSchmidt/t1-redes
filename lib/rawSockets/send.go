package rawsockets

import (
	"errors"
	"fmt"
	"syscall"

	debug "pacman-redes/lib/debug"
)

/*
Envia a mensagem pelo socket especificado. Não possui timeout, retransmissão ou verificação de resposta.
*/
func sendPacket(packet Message) error {
	frame := packet.ToBytes()

	// salvar ultima mensagem enviada
	ServerState.lastSentMessage = packet

	_, err := syscall.Write(ServerState.Sock, frame)

	return err
}

/*
Reenvia os bytes da última mensagem enviada pelo socket especificado. Útil para reenviar ACK/NACK
após receber um pacote duplicado (retransmissão do remetente).
*/
func ResendLastSent() error {
	if ServerState.lastSentMessage.EqualsTo(Message{}) {
		return fmt.Errorf("nenhuma mensagem enviada anteriormente")
	}

	frame := ServerState.lastSentMessage.ToBytes()
	_, err := syscall.Write(ServerState.Sock, frame)
	return err
}

/*
Envia a mensagem pelo socket especificado. Se o tipo da mensagem for ACK, NACK ou Error, é enviada sem aguardar resposta.
Para outros tipos, implementa um mecanSendMessageismo de retransmissão com timeout e limite de tentativas, aguardando alguma resposta.
*/
func SendMessage(packet Message) error {
	if (packet.PacketType == Ack || packet.PacketType == Nack || packet.PacketType == Error) {
		err := sendPacket(packet)
		if err != nil {
			return fmt.Errorf("falha ao enviar mensagem: %w", err)
		}

		return nil
	} else {
		timeoutMillis := initialTimeoutMillis

		for attempt := 1; attempt <= MaxAttempts; attempt++ {
			err := sendPacket(packet)
			if err != nil {
				if !errors.Is(err, syscall.ENETDOWN) {
					return fmt.Errorf("falha ao enviar mensagem: %w", err)
				}
			}

			debug.WriteLog("[MSG] enviado (tentativa %d/%d) => %s\n", attempt, MaxAttempts, packet.String())
			debug.WriteDebug("\t- %d bytes (seq=%d); aguardando ACK por %dms\n", packet.Size(), packet.Sequence, timeoutMillis)

			msg, err := ReceivePacketTWithTimeout(timeoutMillis, Ack)
			
			switch {
			case errors.Is(err, ErrUnexpectedPacketType):
				if msg.PacketType == Nack {
					// reenviar a mensagem, resetando o numero de tentativas
					timeoutMillis = initialTimeoutMillis
					attempt = 0
					continue
				} else if msg.PacketType == Error {
					// retornar para a funcao que chamou para tratar
					if string(msg.Content) == "1" {
						return ErrMissingStorage
					} else {
						return ErrWriteFile
					}
				}
			case errors.Is(err, ErrTimeout):
				debug.WriteLog("\t- sem resposta dentro de %4dms; reenviando...\n", timeoutMillis)
				timeoutMillis = min(timeoutMillis*2, maxTimeoutMillis)
				continue
			case err == nil:
				debug.WriteLog("\t- ack recebido\n")
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
func SendContent(content []byte, pktType PacketT) error {
	if len(content) == 0 {
		msg := CreateMessage(nil, pktType)
		err := SendMessage(msg)
		if err != nil {
			return fmt.Errorf("erro ao enviar conteudo: %w\n", err)
		}
	} else {
		// separar o conteudo em partes e enviar mensagens
		for i := 0; i < len(content); i += maxPacketSize {
			upperBound := i+maxPacketSize
			if upperBound > len(content) {
				upperBound = len(content)
			}

			msg := CreateMessage(content[i:upperBound], pktType)
			err := SendMessage(msg)
			if err != nil {
				return fmt.Errorf("erro ao enviar conteudo: %w\n", err)
			}
		}
	}

	if pktType != EndConn && pktType != End {
		// enviar mensagem de fim
		endMsg := CreateMessage(nil, End)
		err := SendMessage(endMsg)
		if err != nil {
			return fmt.Errorf("erro ao enviar mensagem de fim: %w\n", err)
		}
	}

	return nil
}