package rawsockets

import (
	"errors"
	"fmt"
	"slices"
	"time"

	crc "pacman-redes/lib/crc"
	debug "pacman-redes/lib/debug"
)

const startMarker = 0x7E
const maxPacketSize = 31

var MaxAttempts = 50 // nao eh const para facilitar testes
const initialTimeoutMillis = 500
const maxTimeoutMillis = 4000


var ErrTimeout = errors.New("timeout aguardando mensagem valida")
var ErrInvalidStartMarker = errors.New("marcador de inicio inválido")
var ErrUnexpectedSequence = errors.New("sequência inesperada")
var ErrUnexpectedPacketType = errors.New("tipo de pacote inesperado")
var ErrIgnoredPacket = errors.New("pacote ignorado")
var ErrDuplicatePacket = errors.New("pacote duplicado (retransmissão)")
var ErrInvalidCRC = errors.New("CRC inválido")
var ErrMissingStorage = errors.New("sem espaco de armazenamento")
var ErrWriteFile = errors.New("escrita do arquivo invalida")

// =========== Tipos pacotes ===========

type PacketT uint8

const (
	Ack       PacketT = 0
	Nack      PacketT = 1
	Visualize PacketT = 2
	Init      PacketT = 3
	Data      PacketT = 4
	TxtFile   PacketT = 5
	JpgFile   PacketT = 6
	Mp4File   PacketT = 7
	MoveRight PacketT = 10
	MoveLeft  PacketT = 11
	MoveUp    PacketT = 12
	MoveDown  PacketT = 13
	Error     PacketT = 15
	End       PacketT = 16
)

func (p PacketT) String() string {
	switch p {
	case Ack:
		return "ack"
	case Nack:
		return "nack"
	case Visualize:
		return "visualizacao"
	case Init:
		return "inicializacao"
	case Data:
		return "dados"
	case TxtFile:
		return "arq .txt"
	case JpgFile:
		return "arq .jpg"
	case Mp4File:
		return "arq .mp4"
	case MoveRight:
		return "mov. direita"
	case MoveLeft:
		return "mov. esquerda"
	case MoveUp:
		return "mov. cima"
	case MoveDown:
		return "mov. baixo"
	case Error:
		return "erro"
	case End:
		return "fim"
	}
	return "indefinido"
}


// ========== Estado Servidor ==========

type State struct {
	logQueue        chan string
	SequenceNumber  uint8
	lastReceivedSeq uint8
	hasReceivedPkt  bool
	lastSentMessage Message
}

func (s *State) WriteLog(msg string) {
	s.logQueue <- msg
}

func (s *State) CloseLogWindow() {
	close(s.logQueue)
	time.Sleep(time.Second)
}

func (s *State) addSequence() {
	s.SequenceNumber = (s.SequenceNumber + 1) % 64
}

// necessario para os testes
func (s *State) Reset() {
	s.SequenceNumber = 0
	s.lastReceivedSeq = 0
	s.hasReceivedPkt = false
	s.lastSentMessage = Message{}
}

var ServerState = State{}

// ============= Mensagens =============

type Message struct {
	Content    []byte
	Sequence   uint8
	PacketType PacketT
}

func (m Message) String() string {
	return fmt.Sprintf("tam. dados: %2d | seq: %2d | tipo: %s", len(m.Content), m.Sequence, m.PacketType)
}

// Tamanho total da mensagem na rede em bytes
func (m Message) Size() int {
	return len(m.Content) + 4
}

// Controi um array de bytes representando a mensagem a ser enviada
func (m Message) ToBytes() []byte {
	frame := []byte{
		startMarker, // marcador de inicio
	}

	payload := []byte(m.Content)
	size := uint8(len(payload))

	// garantir que hajam 32 bytes de dados
	if size > maxPacketSize {
		payload = payload[:maxPacketSize] // inclui 0-30 bytes
		size = maxPacketSize
	}

	sequence := m.Sequence & 0x3F

	// segundo byte: primeiros 5 bits de tamanho + 3 bits da sequência
	frame = append(frame, byte((size&0x1F)<<3|(sequence&0x38)>>3))

	// terceiro byte: últimos 3 bits da sequência + 5 bits de tipo
	frame = append(frame, byte((sequence&0x07)<<5)|(uint8(m.PacketType)&0x1F))

	// se houver o identificador de VLAN na possicao 9 e 10 do conteudo,
	// eh necessario adicionar um byte de quebra logo apos
	if size > 10 && (payload[9] == 0x88 || payload[9] == 0x81) {
		newPayload := make([]byte, len(payload))
		copy(newPayload, payload)

		debug.PrintLog("\t- byte p/ quebrar vlan adicionado\n")
		frame = append(frame, slices.Insert(newPayload, 10, 0xff)...)
	} else {
		frame = append(frame, payload...)
	}

	frame = append(frame, crc.CalculateCRC(frame[1:]))

	// tamanho minimo de 15 bytes
	if len(frame) < 15 {
		padding := make([]byte, 15-len(frame))
		frame = append(frame, padding...)
	}

	debug.PrintLog("Mensagem convertida p/ bytes: %x\n", frame)

	return frame
}

func (m Message) EqualsTo(m2 Message) bool {
	if (m.PacketType != m2.PacketType ||
		m.Sequence != m2.Sequence || 
		len(m.Content) != len(m2.Content)) {
		return false
	}

	for i := 0; i < len(m.Content); i++ {
		if m.Content[i] != m2.Content[i] {
			return false
		}
	}

	return true
}

func deformatContent(content []byte) []byte {
	// verificar bytes de VLAN (0x88 e 0xa8)
	if len(content) >= 10 && (content[9] == 0x88 || content[9] == 0x81) {
		var deformattedContent []byte
		deformattedContent = content[:10]

		// adicionar a outra parte do conteudo, se existir
		if len(content) >= 11 {
			deformattedContent = append(deformattedContent, content[11:]...)
		}

		return deformattedContent
	}

	return content
}

/*
Cria uma nova mensagem com o conteúdo e tipo especificados. O número de sequência é
incrementado a cada mensagem criada e lida, garantindo sincronia entre remetente e destinatário.
*/
func CreateMessage(content []byte, PacketT PacketT) Message {
	// incrementar o numero de sequência para a próxima mensagem
	// após a função retornar
	defer ServerState.addSequence()

	return Message{
		Content:    content,
		Sequence:   ServerState.SequenceNumber,
		PacketType: PacketT,
	}
}

/*
Interpreta os bytes do buffer como uma mensagem
*/
func ReadMessage(buf []byte, n int) (Message, error) {
	if n < 4 {
		return Message{}, fmt.Errorf("pacote muito curto")
	}

	if buf[0] != startMarker {
		// descartar pacotes que não começam com o marcador de inicio
		return Message{}, ErrInvalidStartMarker
	}

	size := (buf[1] >> 3)
	if n < int(4+size) {
		return Message{}, fmt.Errorf("pacote muito curto (esperado: %d, recebido: %d)", 4+size, n)
	}

	if n > 15 && (buf[12] == 0x88 || buf[12] == 0x81) {
		size += 1
	}

	bufferUsable := buf[1 : 4+size]
	crcValue := bufferUsable[2+size]

	if !crc.VerifyCRC(bufferUsable[:2+size], crcValue) {
		return Message{}, ErrInvalidCRC
	}

	msg := Message{
		Content:    deformatContent(bufferUsable[2 : 2+size]),
		Sequence:   ((bufferUsable[0] & 0x07) << 3) | (bufferUsable[1] >> 5),
		PacketType: PacketT(bufferUsable[1] & 0x1F),
	}

	debug.PrintLog("Conteudo mensagem (CRC: %d): %v\n", crcValue, msg.Content)

	// detectar retransmissão: sequência igual à última recebida com sucesso
	if ServerState.hasReceivedPkt && msg.Sequence == ServerState.lastReceivedSeq {
		return msg, ErrDuplicatePacket
	}

	// validar numero de sequência
	if msg.Sequence != ServerState.SequenceNumber {
		return Message{}, ErrUnexpectedSequence
	}

	// tudo certo, registrar sequência recebida e incrementar para a próxima mensagem
	ServerState.lastReceivedSeq = msg.Sequence
	ServerState.hasReceivedPkt = true
	ServerState.addSequence()

	return msg, nil
}