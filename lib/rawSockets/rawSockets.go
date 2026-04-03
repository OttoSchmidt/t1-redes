package rawsockets

import (
	"fmt"
	"net"
	"syscall"

	"golang.org/x/sys/unix"

	crc "pacman-redes/lib/crc"
	debug "pacman-redes/lib/debug"
)

const crcPolynomial = 0x07
const ethPAll = 0x0003

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

// Controi um array de bytes representando a mensagem a ser enviada
func buildMessage(content string, id uint8, packetType uint8) []byte {
	frame := []byte{
		0x7E, // marcador de inicio
	}

	payload := []byte(content)
	size := uint8(len(payload))

	if size > 31 {
		payload = payload[:31]
		size = 31
	}

	// segundo byte: primeiros 5 bits de tamanho + 3 bits de ID
	frame = append(frame, byte((size & 0x1F) << 3 | (id & 0x38) >> 3))

	// terceiro byte: últimos 3 bits de ID + 5 bits de tipo
	frame = append(frame, byte((id & 0x07) << 5) | (packetType & 0x1F))

	frame = append(frame, payload...)
	
	frame = append(frame, crc.CalculateCRC(frame[1:]))

	debug.PrintLog("Frame construído: %v\n", frame)

	return frame
}

func SendMessage(sock int, content string, packetType uint8) (int, error) {
	frame := buildMessage(content, 0, packetType)

	n, err := syscall.Write(sock, frame)

	return n, err
}

func ReadMessage(buf []byte, n int) (string, error) {
	if n < 4 {
		return "", fmt.Errorf("pacote muito curto")
	}

	if buf[0] != 0x7E {
		// descartar pacotes que não começam com o marcador de inicio
		return "", fmt.Errorf("marcador de inicio inválido")
	}

	size := (buf[1] >> 3)
	msg := buf[:4+size]
	
	id := ((msg[1] & 0x07) << 3) | (msg[2] >> 5)
	packetType := msg[2] & 0x1F

	if int(size) > n-4 {
		return "", fmt.Errorf("tamanho declarado maior que o recebido")
	}

	content := string(msg[3 : 3+size])
	crcValue := msg[3+size]

	fmt.Printf("Pacote capturado (%d bytes): [tam: %v | seq: %v | tipo: %v | crc: %v]\n", n, size, id, packetType, crcValue)
	debug.PrintLog("Mensagem: %v\n", msg)

	if !crc.VerifyCRC(msg[1:3+size], crcValue) {
		return "", fmt.Errorf("CRC invalido")
	}

	return content, nil
}