package rawsockets

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/sys/unix"

	debug "pacman-redes/lib/debug"
)

const ethPAll = 0x0003
const termExec = "ptyxis"

func init() {
	fmt.Printf("inicializando janela de logs\n")

	// criar canal nomeado
	pipePath := "/tmp/pacman_pipe"
	os.Remove(pipePath)
	syscall.Mkfifo(pipePath, 0666)
	os.Chmod(pipePath, 0666)

	// criar canal
	ServerState.logQueue = make(chan string, 512)

	go logWorker(pipePath)

	// abrir terminal e ler do pipe nomeado. o terminal precisa rodar dentro do usuario logodo no sistema,
	// pois se rodar como root, nao eh possivel criar um outro terminal
	user := os.Getenv("SUDO_USER")
	uid := os.Getenv("SUDO_UID")

	cmd := exec.Command("sudo", "-u", user,
		"env",
		"XDG_RUNTIME_DIR=/run/user/"+uid,
		"ptyxis", "--new-window", "--", "bash", "-c", "cat < "+pipePath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error: %v\nOutput: %s\n", err, string(output))
	}
}

func htons(v uint16) uint16 {
	return (v<<8)&0xff00 | v>>8
}

func CreateSocket(ifaceName string) error {
	// verificar se interface existe
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return fmt.Errorf("falha ao obter interface %s: %w", ifaceName, err)
	}

	// criar file descriptor para socket raw
	protocol := htons(ethPAll)
	sock, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, int(protocol))
	if err != nil {
		return fmt.Errorf("falha ao criar socket raw: %w", err)
	}
	ServerState.Sock = sock

	// vincular o socket a interface
	linkAddr := &syscall.SockaddrLinklayer{
		Protocol: protocol,
		Ifindex:  iface.Index,
	}
	if err := syscall.Bind(sock, linkAddr); err != nil {
		syscall.Close(sock)
		return fmt.Errorf("falha ao vincular socket: %w", err)
	}

	// Habilitar modo promíscuo
	mreq := unix.PacketMreq{
		Ifindex: int32(iface.Index),
		Type:    unix.PACKET_MR_PROMISC,
	}
	if err := unix.SetsockoptPacketMreq(sock, unix.SOL_PACKET, unix.PACKET_ADD_MEMBERSHIP, &mreq); err != nil {
		syscall.Close(sock)
		return fmt.Errorf("falha ao habilitar modo promíscuo: %w", err)
	}

	debug.PrintLog("Socket raw na interface %s (ifindex=%d) com modo promíscuo\n", iface.Name, iface.Index)

	return nil
}

/*
Le as mensagens do canal de logs e escreve num pipe nomeado
*/
func logWorker(pipe string) {
	file, err := os.OpenFile(pipe, os.O_RDWR, os.ModeNamedPipe)
	if err != nil {
		fmt.Printf("erro ao abrir pipe nomeado: %s\n", err)
		return
	}
	defer file.Close()

	// caso a fila acabe, a rotina para aqui.
	// so termina quando o canal eh fechado
	for msg := range ServerState.logQueue {
		file.WriteString(msg)
	}
}