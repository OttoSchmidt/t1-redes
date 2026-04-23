package rawsockets

import (
	"errors"
	"fmt"
	"net"
	"syscall"
	"time"

	"golang.org/x/sys/unix"

	crc "pacman-redes/lib/crc"
	debug "pacman-redes/lib/debug"
)

const ethPAll = 0x0003
const startMarker = 0x7E

const maxAttempts = 50
const initialTimeoutMillis = 500
const maxTimeoutMillis = 4000

const (
	PacketTypeAck  uint8 = 0
	PacketTypeNack uint8 = 1
	PacketTypeData uint8 = 4
)

type Message struct {
	Content    string
	Sequence   uint8
	PacketType uint8
}

func (m Message) String() string {
	return fmt.Sprintf("Tamanho: %d, Sequencia: %d, Tipo: %d", len(m.Content), m.Sequence, m.PacketType)
}

// Controi um array de bytes representando a mensagem a ser enviada
func (m Message) toBytes() []byte {
	frame := []byte{
		startMarker, // marcador de inicio
	}

	payload := []byte(m.Content)
	size := uint8(len(payload))

	// garantir que hajam 32 bytes de dados
	if size > 31 {
		payload = payload[:31]
		size = 31
	}

	sequence := m.Sequence & 0x3F

	// segundo byte: primeiros 5 bits de tamanho + 3 bits da sequência
	frame = append(frame, byte((size&0x1F)<<3|(sequence&0x38)>>3))

	// terceiro byte: últimos 3 bits da sequência + 5 bits de tipo
	frame = append(frame, byte((sequence&0x07)<<5)|(m.PacketType&0x1F))

	frame = append(frame, payload...)

	frame = append(frame, crc.CalculateCRC(frame[1:]))

	// tamanho minimo de 15 bytes
	if (len(frame) < 15) {
		padding := make([]byte, 15-len(frame))
		frame = append(frame, padding...)
	}

	debug.PrintLog("Mensagem convertida p/ bytes: %v\n", frame)

	return frame
}

type State struct {
	SequenceNumber    uint8
	LastSentMessage   Message
}

func (s *State) addSequence() {
	s.SequenceNumber = (s.SequenceNumber + 1) % 64
}

var ErrTimeout = errors.New("timeout aguardando mensagem valida")
var ErrInvalidStartMarker = errors.New("marcador de inicio inválido")

var ServerState = State{}

func htons(v uint16) uint16 {
	return (v<<8)&0xff00 | v>>8
}

func CreateRawSocket(ifaceName string) (int, error) {
	// verificar se interface existe
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return 0, fmt.Errorf("falha ao obter interface %s: %w", ifaceName, err)
	}

	// criar file descriptor para socket raw
	protocol := htons(ethPAll)
	sock, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, int(protocol))
	if err != nil {
		return 0, fmt.Errorf("falha ao criar socket raw: %w", err)
	}

	// vincular o socket a interface
	linkAddr := &syscall.SockaddrLinklayer{
		Protocol: protocol,
		Ifindex:  iface.Index,
	}
	if err := syscall.Bind(sock, linkAddr); err != nil {
		syscall.Close(sock)
		return 0, fmt.Errorf("falha ao vincular socket: %w", err)
	}

	// Habilitar modo promíscuo
	mreq := unix.PacketMreq{
		Ifindex: int32(iface.Index),
		Type:    unix.PACKET_MR_PROMISC,
	}
	if err := unix.SetsockoptPacketMreq(sock, unix.SOL_PACKET, unix.PACKET_ADD_MEMBERSHIP, &mreq); err != nil {
		syscall.Close(sock)
		return 0, fmt.Errorf("falha ao habilitar modo promíscuo: %w", err)
	}

	debug.PrintLog("Socket raw na interface %s (ifindex=%d) com modo promíscuo\n", iface.Name, iface.Index)

	return sock, nil
}




func CreateMessage(content string, packetType uint8) Message {
	// incrementar o numero de sequência para a próxima mensagem
	// após a função retornar
	defer ServerState.addSequence()

	return Message{
		Content: content,
		Sequence: ServerState.SequenceNumber,
		PacketType: packetType,
	}
}

func SendMessage(sock int, packet Message) (int, error) {
	frame := packet.toBytes()

	n, err := syscall.Write(sock, frame)

	return n, err
}

func AttemptSendMessage(sock int, packet Message) error {
	timeoutMillis := initialTimeoutMillis

	err := error(nil)
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		n, err := SendMessage(sock, packet)
		if err != nil {
			return fmt.Errorf("falha ao enviar mensagem: %w", err)
		}

		fmt.Printf("Tentativa %d/%d: enviado %d bytes (seq=%d); aguardando ACK por %dms\n", attempt, maxAttempts, n, packet.Sequence, timeoutMillis)

		ack, err := ReceivePacketTypeWithTimeout(sock, timeoutMillis, PacketTypeAck)
		if err == nil {
			fmt.Printf("ACK recebido: %s\n", ack.Content)
			return nil
		}

		if errors.Is(err, ErrTimeout) {
			fmt.Printf("Sem ACK dentro de %dms; reenviando...\n", timeoutMillis)
			timeoutMillis = min(timeoutMillis*2, maxTimeoutMillis)
			continue
		}

		return fmt.Errorf("erro ao aguardar ACK: %w", err)
	}

	return fmt.Errorf("falha ao obter ACK: limite de tentativas atingido: %w", err)
}

func ReadPacket(buf []byte) (Message, error) {
	if len(buf) < 4 {
		return Message{}, fmt.Errorf("pacote muito curto")
	}

	if buf[0] != startMarker {
		// descartar pacotes que não começam com o marcador de inicio
		return Message{}, ErrInvalidStartMarker
	}

	size := (buf[1] >> 3)
	bufferUsable := buf[1:4+size]
	crcValue := bufferUsable[2+size]

	msg := Message{
		Content:    string(bufferUsable[2 : 2 + size]),
		Sequence:   ((bufferUsable[0] & 0x07) << 3) | (bufferUsable[1] >> 5),
		PacketType: bufferUsable[1] & 0x1F,
	}

	// validar numero de sequência
	if msg.Sequence != ServerState.SequenceNumber {
		return Message{}, fmt.Errorf("sequencia inesperada: esperado %d, recebido %d", ServerState.SequenceNumber, msg.Sequence)
	}
	ServerState.addSequence()

	fmt.Printf("Mensagem capturada (CRC: %d): %s\n", crcValue, msg.String())
	debug.PrintLog("Conteudo mensagem: %v\n", msg.Content)

	if !crc.VerifyCRC(bufferUsable[:2+size], crcValue) {
		return Message{}, fmt.Errorf("CRC invalido")
	}

	return msg, nil
}

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

		_, addr, err := syscall.Recvfrom(sock, buf, 0)
		if err != nil {
			if errors.Is(err, syscall.EAGAIN) || errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EINTR) {
				continue
			}

			return Message{}, fmt.Errorf("falha no recvfrom: %w", err)
		}

		if llAddr, ok := addr.(*syscall.SockaddrLinklayer); ok && llAddr.Pkttype == syscall.PACKET_OUTGOING {
			continue
		}

		msg, err := ReadPacket(buf)
		if err != nil {
			debug.PrintLog("Pacote ignorado durante espera: %v\n", err)
			continue
		}

		return msg, nil
	}
}

func ReceivePacketTypeWithTimeout(sock int, timeoutMillis int, expectedType uint8) (Message, error) {
	deadline := time.Now().Add(time.Duration(timeoutMillis) * time.Millisecond)

	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return Message{}, ErrTimeout
		}

		msg, err := ReceivePacketWithTimeout(sock, int(remaining/time.Millisecond))
		if err != nil {
			if errors.Is(err, ErrTimeout) {
				return Message{}, ErrTimeout
			}

			return Message{}, err
		}

		if msg.PacketType != expectedType {
			debug.PrintLog("Pacote tipo %d ignorado; aguardando tipo %d\n", msg.PacketType, expectedType)
			continue
		}

		return msg, nil
	}
}
