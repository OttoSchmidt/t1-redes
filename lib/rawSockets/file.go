package rawsockets

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/sys/unix"

	debug "pacman-redes/lib/debug"
)

func VerifyFileViability(id int, tam uint, fileType PacketT) (*os.File, error) {
	var fileExt string
	switch fileType {
	case TxtFile:
		fileExt = ".txt"
	case JpgFile:
		fileExt = ".jpg"
	case Mp4File:
		fileExt = ".mp4"
	default:
		return nil, fmt.Errorf("tipo de arquivo invalido: %d", fileType)
	}

	// verificar permissoes de arquivo temporario
	fileName := fmt.Sprintf("pacman%d-*%s", id, fileExt)
	tmpFile, err := os.CreateTemp("/tmp", fileName)
	if err != nil {
		os.Remove(tmpFile.Name())
		return nil, err
	}
	fileName = tmpFile.Name()

	// verificar se ha espaco disponivel
	var stat unix.Statfs_t
	unix.Statfs("/tmp", &stat) //recuperar informacoes de /tmp 

	// blocos disponiveis * tamanho do bloco
	tamAvailable := stat.Bavail * uint64(stat.Bsize)
	if tamAvailable < uint64(tam) {
		debug.PrintLog("Espaco disponivel: %d bytes; necessario: %d bytes\n", tamAvailable, tam)
		os.Remove(fileName)
		return nil, ErrMissingStorage
	}

	// renomear arquivo
	newFileName := fmt.Sprintf("/tmp/%d%s", id, fileExt)
	tmpFile.Close()
	os.Rename(fileName, newFileName)
	tmpFile, err = os.OpenFile(newFileName, os.O_RDWR, 0666)
	if err != nil {
		os.Remove(fileName)
		os.Remove(newFileName)
		return nil, fmt.Errorf("nao foi possivel renomar arquivo")
	}

	return tmpFile, nil
}

func ParseFileHeader(content []byte) (id int, tam uint, err error) {
	debug.PrintLog("cabecalho arquivo recebido: %s\n", string(content))

	_, err = fmt.Sscanf(string(content), "%c%d", &id, &tam)
	if err != nil {
		return 0, 0, fmt.Errorf("formato de cabecalho de arquivo invalido: %w", err)
	}

	return id, tam, nil
}

func OpenDefaultFileHandler(file string) error {
	cmd := exec.Command("xdg-open", file)
	return cmd.Start()
}

func ReceiveFile(sock int, file *os.File, tam uint) (string, error) {
	receivedBytes := uint(0)
	fileBuffer := make([]byte, tam)
	buf := make([]byte, 40)

	for receivedBytes < tam {
		msg, err := ReceivePacket(sock, buf)
		if err != nil {
			debug.PrintLog("Erro ao receber pacote de arquivo: %v\n", err)

			if errors.Is(err, ErrInvalidCRC) {
				// enviar NACK para solicitar retransmissão
				nackMsg := CreateMessage(nil, Nack)
				if sendErr := SendMessage(sock, nackMsg); sendErr != nil {
					debug.PrintLog("Erro ao enviar NACK para pacote com CRC invalido: %v\n", sendErr)
				}
			}
			continue
		}

		ackMsg := CreateMessage(nil, Ack)
		if sendErr := SendMessage(sock, ackMsg); sendErr != nil {
			debug.PrintLog("Erro ao enviar ACK para pacote recebido: %v\n", sendErr)
			continue
		}

		if msg.PacketType != Data {
			os.Remove(file.Name())
			return "", ErrUnexpectedPacketType
		}

		// escrever os bytes recebidos no buffer de arquivo
		copy(fileBuffer[receivedBytes:], msg.Content)
		receivedBytes += uint(len(msg.Content))
		debug.PrintLog("ARQ: falta %d de %d bytes\n", tam - receivedBytes, tam)
	}

	// escrever o buffer de arquivo no arquivo temporario
	_, err := file.Write(fileBuffer)
	if err != nil {
		os.Remove(file.Name())
		return "", err
	}

	fileName := file.Name()
	file.Close()

	os.Chmod(fileName, 0644)

	return fileName, nil
}

func SendFile(sock int, id int, file *os.File) error {
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("falha ao obter informacoes do arquivo: %w", err)
	}

	// determinar tipo do arquivo (txt, jpg, mp4)
	var fileType PacketT
	fileExt := filepath.Ext(fileInfo.Name())
	switch fileExt {
	case ".txt":
		fileType = TxtFile
	case ".jpg", ".jpeg":
		fileType = JpgFile
	case ".mp4":
		fileType = Mp4File
	default:
		return fmt.Errorf("tipo de arquivo nao suportado: %s", fileExt)
	}

	// ler arquivo
	content := make([]byte, fileInfo.Size())
	_, err = file.Read(content)
	if err != nil {
		return fmt.Errorf("falha ao ler arquivo: %w", err)
	}

	// enviar pacote cabecalho
	msg := CreateMessage([]byte(fmt.Sprintf("%c%d", id&0xff, fileInfo.Size())), fileType)
	err = SendMessage(sock, msg)
	if err != nil {
		return fmt.Errorf("falha ao enviar cabecalho do arquivo: %w", err)
	}

	// enviar o conteudo do arquivo
	return SendContent(sock, content, Data)
}