package pacman

import (
	rawsockets "pacman-redes/lib/rawSockets"
)


func (p *Position) detectCollision(p2 *Position) bool {
	return p.x == p2.x && p.y == p2.y
}

/*
func (g *Ghost) movement() {
	switch g.ent.symbol {
		case 
	}
}
*/

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


	// detectar colisao
	

	return nil
}