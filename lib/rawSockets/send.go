package rawsockets

import (
	"errors"
	"fmt"
	"syscall"
)

/*
Envia a mensagem pelo socket especificado. Não possui timeout, retransmissão ou verificação de resposta.
*/
func sendPacket(sock int, packet Message) error {
	frame := packet.toBytes()

	_, err := syscall.Write(sock, frame)

	return err
}

/*
Envia a mensagem pelo socket especificado. Se o tipo da mensagem for ACK, NACK ou Error, é enviada sem aguardar resposta.
Para outros tipos, implementa um mecanismo de retransmissão com timeout e limite de tentativas, aguardando alguma resposta.
*/
func SendMessage(sock int, packet Message) error {
	if (packet.PacketType == PacketTypeAck || packet.PacketType == PacketTypeNack || packet.PacketType == PacketTypeError) {
		err := sendPacket(sock, packet)
		if err != nil {
			return fmt.Errorf("falha ao enviar mensagem: %w", err)
		}

		return nil
	} else {
		timeoutMillis := initialTimeoutMillis

		for attempt := 1; attempt <= maxAttempts; attempt++ {
			err := sendPacket(sock, packet)
			if err != nil {
				return fmt.Errorf("falha ao enviar mensagem: %w", err)
			}

			fmt.Printf("Tentativa %d/%d: enviado %d bytes (seq=%d); aguardando ACK por %dms\n", attempt, maxAttempts, packet.Size(), packet.Sequence, timeoutMillis)

			msg, err := ReceivePacketTypeWithTimeout(sock, timeoutMillis, PacketTypeAck)
			
			switch {
			case errors.Is(err, ErrUnexpectedPacketType):
				if msg.PacketType == PacketTypeNack {
					// reenviar a mensagem, resetando o numero de tentativas
					attempt = 1
					continue
				}
			case errors.Is(err, ErrTimeout):
				fmt.Printf("Sem resposta dentro de %dms; reenviando...\n", timeoutMillis)
				timeoutMillis = min(timeoutMillis*2, maxTimeoutMillis)
				continue
			case err == nil:
				fmt.Printf("ACK recebido\n")
				return nil
			default:
				return fmt.Errorf("erro ao aguardar ACK: %w", err)
			}
		}

		return fmt.Errorf("falha ao obter ACK: limite de tentativas atingido")
	}
}