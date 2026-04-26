package rawsockets_test

import (
	"errors"
	"fmt"
	"syscall"
	"testing"
	"time"

	rawsockets "pacman-redes/lib/rawSockets"
)

func createTestSocket(ifaceName string) (int, error) {
	return rawsockets.CreateSocket(ifaceName)
}

func TestSuccessACK(t *testing.T) {
	defer rawsockets.ServerState.Reset()

    // canais para sincronização
    serverReady := make(chan bool)
    serverDone := make(chan error, 1)
	// fluxo do servidor
    go func() {
        sockServer, err := createTestSocket("lo")
        if err != nil {
            serverDone <- fmt.Errorf("erro ao abrir socket do servidor: %w", err)
            return
        }
        defer syscall.Close(sockServer)

		// sinaliza que o servidor está escutando
        serverReady <- true

		// Aumentar o timeout do servidor e inicializar o log
		_, err = rawsockets.ReceivePacketWithTimeout(sockServer, 3000)
		if err == nil {			
			ackMsg := rawsockets.Message{
				Content:    "",
				Sequence:   rawsockets.ServerState.SequenceNumber,
				PacketType: rawsockets.Ack,
			}
			
			err = rawsockets.SendMessage(sockServer, ackMsg)
		}
        
        serverDone <- err
    }()

    // fluxo do cliente:
    // espera até que o servidor esteja ativo
    <-serverReady

    // cliente abre o próprio socket
    sockClient, err := createTestSocket("lo")
    if err != nil {
        t.Fatalf("erro ao abrir socket do cliente: %v", err)
    }
    defer syscall.Close(sockClient)

	// cliente envia mensagem de teste p/ servidor
    msgData := rawsockets.CreateMessage("Teste ACK Real", rawsockets.Data)
	rawsockets.ServerState.Reset()
    err = rawsockets.SendMessage(sockClient, msgData)
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
		sockServer, err := createTestSocket("lo")
		if err != nil {
			serverDone <- fmt.Errorf("erro ao abrir socket do servidor: %w", err)
			return
		}
		defer syscall.Close(sockServer)

		// sinaliza que o servidor está escutando
		serverReady <- true

		// esperar por pelo menos um timeout do cliente
		time.Sleep(520 * time.Millisecond)

		_, err = rawsockets.ReceivePacketWithTimeout(sockServer, 1000)
		serverDone <- err
	}()

	// espera até que o servidor esteja ativo
	<-serverReady

	sockClient, err := createTestSocket("lo")
	if err != nil {
		t.Fatalf("erro ao abrir socket do cliente: %v", err)
	}
	defer syscall.Close(sockClient)

	msgData := rawsockets.CreateMessage("Teste ACK com timeout", rawsockets.Data)
	rawsockets.ServerState.Reset()
	
	// Executa SendMessage do cliente separadamente p/ não travar o select do teste
	go func() {
		_ = rawsockets.SendMessage(sockClient, msgData)
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

	clientSock, err := createTestSocket("lo")
	if err != nil {
		t.Fatalf("erro ao abrir socket do cliente: %v", err)
	}
	defer syscall.Close(clientSock)

	msg := rawsockets.CreateMessage("Teste de máximo de tentativas atingido", rawsockets.Data)
	rawsockets.ServerState.Reset()

	err = rawsockets.SendMessage(clientSock, msg)
	if err == nil {
		t.Fatal("esperava erro após atingir o máximo de tentativas, mas envio foi bem-sucedido")
	}
}