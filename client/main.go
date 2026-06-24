package main

import (
	"fmt"
	"log"
	"os"
	"syscall"
	"time"

	tea "charm.land/bubbletea/v2"

	debug "pacman-redes/lib/debug"
	pacman "pacman-redes/lib/pacman"
	rawsockets "pacman-redes/lib/rawSockets"
)

type model string
type tickMsg time.Time

// gerar ticks a cada segundo
func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Init() tea.Cmd {
	return tick()
}

func (m model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyPressMsg:
		key := msg.String()
		switch key {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "w", "a", "s", "d":
			sendMovement(key)
			return getNewMap(m)
		}
	}

	return m, nil
}

func (m model) View() tea.View {
	v := tea.NewView(string(m))
	v.AltScreen = true
	return v
}

func getNewMap(oldModel model) (model, tea.Cmd) {
	buf := make([]byte, 256)

	content, packetType, err := rawsockets.ReceiveContent(buf)
	if err != nil {
		debug.WriteDebug("\t- erro ao receber conteudo:\n\t- %v\n", err)
	}

	switch packetType {
	case rawsockets.Init, rawsockets.Visualize:
		grid, center, radius := pacman.GridFromBytes(content)
		return model(grid.ToString(center, radius)), nil
	case rawsockets.EndConn:
		return oldModel, tea.Quit
	}

	return oldModel, nil
}

func sendMovement(direcao string) {
	var keyType rawsockets.PacketT
	switch direcao {
	case "w":
		keyType = rawsockets.MoveUp
	case "a":
		keyType = rawsockets.MoveLeft
	case "s":
		keyType = rawsockets.MoveDown
	case "d":
		keyType = rawsockets.MoveRight
	default:
		return
	}

	// enviar direcao
	err := rawsockets.SendContent(nil, keyType)
	if (err != nil) {
		debug.WriteLog("\t- erro ao enviar movimento ao servidor: %s\n", err.Error())
	}
}

func main() {
	// interface padrao: loopback. para usar outra interface,
	// passe o nome como argumento,como: eth0, enp3s0...
	ifaceName := "lo"
	if len(os.Args) > 1 {
		ifaceName = os.Args[1]
	}

	err := rawsockets.CreateSocket(ifaceName)
	defer syscall.Close(rawsockets.ServerState.Sock)
	if err != nil {
		panic(err)
	}

	initMap, _ := getNewMap("")

	p := tea.NewProgram(model(initMap))
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("encerrando...\n");
}