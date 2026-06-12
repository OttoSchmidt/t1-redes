package pacman

import (
	"fmt"
	"os"
	rawsockets "pacman-redes/lib/rawSockets"
	"slices"
)


func (p *Position) detectCollision(p2 *Position) bool {
	return p.x == p2.x && p.y == p2.y
}

/*
func (g *Ghost) movement() {
	switch g.ent.symbol {
		case 
	}
}*/

func (gs *GameState) MovePlayer(pkt rawsockets.PacketT) error {
	// mover player
	newPlayerPos := gs.GameMap.pacman.ent.pos 
	switch pkt {
	case rawsockets.MoveDown:
		if newPlayerPos.y < 39 {
			newPlayerPos.y++
		}
	case rawsockets.MoveUp:
		if newPlayerPos.y > 0 {
			newPlayerPos.y--
		}
	case rawsockets.MoveLeft:
		if newPlayerPos.x > 0 {
			newPlayerPos.x--
		}
	case rawsockets.MoveRight:
		if newPlayerPos.x < 39 {
			newPlayerPos.x++
		}
	}
	
	// verificar colisao com parede e limite do mapa
	if gs.GameMap.grid[newPlayerPos.y][newPlayerPos.x] != 'X' {
		gs.GameMap.pacman.ent.pos.x = newPlayerPos.x
		gs.GameMap.pacman.ent.pos.y = newPlayerPos.y
	}

	// mover fantasmas


	// detectar colisao com fantasmas
	for _, g := range gs.GameMap.ghosts {
		if gs.GameMap.pacman.ent.pos.detectCollision(&g.ent.pos) {
			// enviar arquivo
			file, err := os.OpenFile("./files/jumpscare.jpg", os.O_RDONLY, 0)
			if err != nil {
				panic(err)
			}
			defer file.Close()
			err = rawsockets.SendFile(7, file)
			if err != nil {
				rawsockets.ServerState.WriteLog(fmt.Sprintf("[ERRO] %s\n", err.Error()))
				break
			}

			break
		}
	}

	// detectar colisao com moedas
	for i, c := range gs.GameMap.coins {
		if gs.GameMap.pacman.ent.pos.detectCollision(&c.ent.pos) {
			// enviar arquivo da moeda
			file, err := os.OpenFile(c.fileName, os.O_RDONLY, 0)
			if err != nil {
				panic(err)
			}
			defer file.Close()
			err = rawsockets.SendFile(int(c.ent.symbol), file)
			if err != nil {
				rawsockets.ServerState.WriteLog(fmt.Sprintf("[ERRO] %s\n", err.Error()))
				break
			}

			// remover moeda
			gs.GameMap.coins = slices.Delete(gs.GameMap.coins, i, i+1)

			break
		}
	}
	

	return nil
}