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
		return Message{}, fmt.Errorf("pacote ignorado")
	}

	msg, err := ReadMessage(buf, n)
	if err != nil {
		if err != ErrInvalidStartMarker {
			debug.PrintLog("Erro ao ler mensagem: %v\n", err)
		}
			
		return Message{}, err
	}

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
			continue
		}

		return msg, err
	}
}

/*
Recebe um pacote do socket, aguardando um tempo máximo especificado (timeout) e filtrando por um tipo especifico.
*/
func ReceivePacketTypeWithTimeout(sock int, timeoutMillis int, expectedType uint8) (Message, error) {
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
