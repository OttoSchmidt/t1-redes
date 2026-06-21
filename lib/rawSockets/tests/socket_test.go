package rawsockets_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	rawsockets "pacman-redes/lib/rawSockets"
)

func createTestSocket(ifaceName string) (error) {
	return rawsockets.CreateSocket(ifaceName)
}

func TestSuccessACK(t *testing.T) {
	defer rawsockets.ServerState.Reset()

    // canais para sincronização
    serverReady := make(chan bool)
    serverDone := make(chan error, 1)
	// fluxo do servidor
    go func() {
        err := createTestSocket("lo")
        if err != nil {
            serverDone <- fmt.Errorf("erro ao abrir socket do servidor: %w", err)
            return
        }

		// sinaliza que o servidor está escutando
        serverReady <- true

		// Aumentar o timeout do servidor e inicializar o log
		_, err = rawsockets.ReceivePacketWithTimeout(3000)
		if err == nil {			
			ackMsg := rawsockets.Message{
				Content:    nil,
				Sequence:   rawsockets.ServerState.SequenceNumber,
				PacketType: rawsockets.Ack,
			}
			
			err = rawsockets.SendMessage(ackMsg)
		}
        
        serverDone <- err
    }()

    // fluxo do cliente:
    // espera até que o servidor esteja ativo
    <-serverReady

    // cliente abre o próprio socket
    err := createTestSocket("lo")
    if err != nil {
        t.Fatalf("erro ao abrir socket do cliente: %v", err)
    }

	// cliente envia mensagem de teste p/ servidor
    msgData := rawsockets.CreateMessage([]byte("Teste ACK Real"), rawsockets.Data)
	rawsockets.ServerState.Reset()
    err = rawsockets.SendMessage(msgData)
    if err != nil {
        t.Fatalf("esperava sucesso no envio, mas estourou timeout/erro: %v", err)
    }

    // finalizacao e validacao geral
    select {
    case err := <-serverDone:
        if err != nil {
            t.Fatalf("a rotina do servidor reportou um erro: %v", err)
        }
    case <-time.After(2 * time.Second):
        t.Fatal("timeout do teste: o servidor nao respondeu")
    }
}

func TestTimeoutACK(t *testing.T) {
	defer rawsockets.ServerState.Reset()

    serverReady := make(chan bool)
    serverDone := make(chan error, 1)

	go func() {
		err := createTestSocket("lo")
		if err != nil {
			serverDone <- fmt.Errorf("erro ao abrir socket do servidor: %w", err)
			return
		}

		// sinaliza que o servidor está escutando
		serverReady <- true

		// esperar por pelo menos um timeout do cliente
		time.Sleep(520 * time.Millisecond)

		_, err = rawsockets.ReceivePacketWithTimeout(1000)
		serverDone <- err
	}()

	// espera até que o servidor esteja ativo
	<-serverReady

	err := createTestSocket("lo")
	if err != nil {
		t.Fatalf("erro ao abrir socket do cliente: %v", err)
	}

	msgData := rawsockets.CreateMessage([]byte("Teste ACK com timeout"), rawsockets.Data)
	rawsockets.ServerState.Reset()
	
	// Executa SendMessage do cliente separadamente p/ não travar o select do teste
	go func() {
		_ = rawsockets.SendMessage(msgData)
	}()

	select {
	case err := <-serverDone:
		if !errors.Is(err, rawsockets.ErrTimeout) {
			t.Fatalf("esperava erro do tipo ErrTimeout, mas obteve: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout do teste: o servidor nao respondeu")
	}
}

func TestMaxRetriesReached(t *testing.T) {
	defer rawsockets.ServerState.Reset()

	// Reduz o máximo de tentativas só pra esse teste rodar rápido!
	originalMaxAttempts := rawsockets.MaxAttempts
	rawsockets.MaxAttempts = 3
	defer func() { rawsockets.MaxAttempts = originalMaxAttempts }()

	err := createTestSocket("lo")
	if err != nil {
		t.Fatalf("erro ao abrir socket do cliente: %v", err)
	}

	msg := rawsockets.CreateMessage([]byte("Teste de máximo de tentativas atingido"), rawsockets.Data)
	rawsockets.ServerState.Reset()

	err = rawsockets.SendMessage(msg)
	if err == nil {
		t.Fatal("esperava erro após atingir o máximo de tentativas, mas envio foi bem-sucedido")
	}
}