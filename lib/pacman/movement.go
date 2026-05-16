package pacman

import rawsockets "pacman-redes/lib/rawSockets"


func (e *Entity) detectCollision(e2 *Entity) bool {
	return e.pos.x == e2.pos.x && e.pos.y == e2.pos.y
}

func (g *Ghost) movement() {

}

func (gs *GameState) MovePlayer(rawsockets.PacketT) error {
	// mover fantasmas

	// mover player

	// detectar colisao
	
	return nil
}