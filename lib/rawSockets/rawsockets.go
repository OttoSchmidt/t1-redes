package rawsockets

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/sys/unix"

	debug "pacman-redes/lib/debug"
)

const ethPAll = 0x0003
const termExec = "ptyxis"
const msgLogsFile = "/tmp/pacman-msg.log"

var pipeReader *io.PipeReader
var pipeWriter *io.PipeWriter

func init() {
	fmt.Printf("inicializando janela de logs\n")
	pipeReader, pipeWriter = io.Pipe()
	go LogWindow()
}

func htons(v uint16) uint16 {
	return (v<<8)&0xff00 | v>>8
}

func CreateSocket(ifaceName string) (int, error) {
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

func LogWindow() {
	// criar arquivo
	file, err := os.Create(msgLogsFile)
	if err != nil {
		panic(err)
	}

	go func() {
		// abrir novo terminal p/ ler do arquivo de log
		cmd := exec.Command(termExec, "--", "tail", "-f", msgLogsFile)
		cmd.Run()
	}()

	for ;; {
		_, err = io.Copy(file, pipeReader)
		if err != nil {
			panic(err)
		}
	}
}