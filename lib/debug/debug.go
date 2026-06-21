package debug

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
)

var logQueue chan string

func init() {
	fmt.Printf("inicializando janela de logs\n")
	if Debug {
		fmt.Printf("modo debug ativado\n")
	}

	// criar canal nomeado
	pipePath := "/tmp/pacman_pipe"
	os.Remove(pipePath)
	syscall.Mkfifo(pipePath, 0666)
	os.Chmod(pipePath, 0666)

	// criar canal
	logQueue = make(chan string, 512)

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

func CloseLogWindow() {
	close(logQueue)
	time.Sleep(time.Second)
}

// Escreve na janela de logs
func WriteLog(format string, a ...interface{}) {
	logQueue <- fmt.Sprintf(format, a...)
}

// Escreve na janela de logs somente se o modo DEBUG estiver habilitado
func WriteDebug(format string, a ...interface{}) {
	if Debug {
		logQueue <- fmt.Sprintf("[DEBUG] " + format, a...)
	}
}

// Le as mensagens do canal de logs e escreve num pipe nomeado
func logWorker(pipe string) {
	file, err := os.OpenFile(pipe, os.O_RDWR, os.ModeNamedPipe)
	if err != nil {
		fmt.Printf("erro ao abrir pipe nomeado: %s\n", err)
		return
	}
	defer file.Close()

	// caso a fila acabe, a rotina para aqui.
	// so termina quando o canal eh fechado
	for msg := range logQueue {
		file.WriteString(msg)
	}
}