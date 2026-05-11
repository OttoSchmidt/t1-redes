package rawsockets_test

import (
	"errors"
	"fmt"
	"testing"

	crc "pacman-redes/lib/crc"
	rawsockets "pacman-redes/lib/rawSockets"
)

func TestSequenceIncrement(t *testing.T) {
	defer rawsockets.ServerState.Reset()

	num := 5
	messages := make([]rawsockets.Message, num)
	for i := 0; i < num; i++ {
		messages[i] = rawsockets.CreateMessage([]byte(fmt.Sprintf("Mensagem %d", i)), rawsockets.Data)
	}

	valid := true
	for i := 0; i < num; i++ {
		expectedSeq := uint8(i % 64)
		if messages[i].Sequence != expectedSeq {
			t.Fatalf("Teste falhou: mensagem %d tem sequência %d, esperado %d\n", i, messages[i].Sequence, expectedSeq)
			valid = false
		}
	}

	if !valid {
		t.Fatal("Teste de incremento de sequência falhou.")
	}
}

func TestReadMessagePacket(t *testing.T) {
	defer rawsockets.ServerState.Reset()

	// msg: tam: 3, seq: 0, tipo: 4, conteudo: 'ola'
	packet := []byte{0x7E, 0x18, 0x04, 0x6F, 0x6C, 0x61, 0xD8}

	message, err := rawsockets.ReadMessage(packet, len(packet))
	if err != nil {
		t.Fatalf("Teste de leitura de pacote de mensagem falhou: %v\n", err)
	}

	expectedContent := "ola"
	if string(message.Content) != expectedContent {
		t.Fatalf("Teste de leitura de pacote de mensagem falhou: conteúdo esperado '%s', obtido '%s'\n", expectedContent, message.Content)
	}
}

func TestInvalidCRCTest(t *testing.T) {
	defer rawsockets.ServerState.Reset()

	msg := rawsockets.CreateMessage([]byte("Teste de CRC inválido"), rawsockets.Data)
	packet := msg.ToBytes()
	rawsockets.ServerState.Reset()

	// altera ultimo byte da parte de dados
	packet[len(packet)-2] ^= 0xFF

	_, err := rawsockets.ReadMessage(packet, len(packet))
	if err == nil {
		t.Fatal("Teste de CRC inválido falhou: mensagem lida sem erro, apesar do CRC inválido")
	}

	if !errors.Is(err, rawsockets.ErrInvalidCRC) {
		t.Fatalf("Teste de CRC inválido falhou: erro esperado do tipo ErrInvalidCRC, obtido: %v\n", err)
	}
}

func TestDuplicateSequence(t *testing.T) {
	defer rawsockets.ServerState.Reset()

	msg := rawsockets.CreateMessage([]byte("Teste de sequência duplicada"), rawsockets.Data)
	packet := msg.ToBytes()
	rawsockets.ServerState.Reset()

	// simula recebimento da mesma mensagem duas vezes
	_, err1 := rawsockets.ReadMessage(packet, len(packet))
	_, err2 := rawsockets.ReadMessage(packet, len(packet))

	if err1 != nil {
		t.Fatalf("Teste de sequência duplicada falhou: erro ao ler primeira mensagem: %v\n", err1)
	}

	if !errors.Is(err2, rawsockets.ErrDuplicatePacket) {
		t.Fatalf("Teste de sequência duplicada falhou: erro esperado do tipo ErrDuplicatePacket para a segunda leitura, obtido: %v\n", err2)
	}
}

func TestDuplicateSequenceWithInvalidCRC(t *testing.T) {
	defer rawsockets.ServerState.Reset()

	msg := rawsockets.CreateMessage([]byte("abc"), rawsockets.Data)
	packet := msg.ToBytes()
	rawsockets.ServerState.Reset()

	// primeira leitura válida, para registrar o último pacote recebido
	if _, err := rawsockets.ReadMessage(packet, len(packet)); err != nil {
		t.Fatalf("Teste falhou: erro ao ler primeira mensagem: %v\n", err)
	}

	// corrompe o CRC exato do pacote duplicado
	crcIndex := 3 + len(msg.Content)
	packet[crcIndex] ^= 0xFF

	_, err := rawsockets.ReadMessage(packet, len(packet))
	if !errors.Is(err, rawsockets.ErrInvalidCRC) {
		t.Fatalf("Teste falhou: esperado ErrInvalidCRC para pacote duplicado corrompido, obtido: %v\n", err)
	}
}

func TestFutureSequence(t *testing.T) {
	defer rawsockets.ServerState.Reset()

	msg := rawsockets.CreateMessage([]byte("Teste de sequência futura"), rawsockets.Data)
	packet := msg.ToBytes()
	rawsockets.ServerState.Reset()

	// simula recebimento de uma mensagem com sequência futura (não a próxima esperada)
	nextSequence := (msg.Sequence + 2) % 64
	packet[1] = (packet[1] & 0xF8) | (nextSequence >> 3)
	packet[2] = (nextSequence<<5)&0xE0 | (packet[2] & 0x1F)

	// após alterar cabeçalho, recalcula CRC para isolar a validação de sequência
	size := len(msg.Content)
	crcIndex := 3 + size
	packet[crcIndex] = crc.CalculateCRC(packet[1 : 3+size])

	_, err := rawsockets.ReadMessage(packet, len(packet))
	if err == nil {
		t.Fatal("Teste de sequência futura falhou: mensagem lida sem erro, apesar da sequência inesperada")
	}

	if !errors.Is(err, rawsockets.ErrUnexpectedSequence) {
		t.Fatalf("Teste de sequência futura falhou: erro esperado do tipo ErrUnexpectedSequence, obtido: %v\n", err)
	}
}
