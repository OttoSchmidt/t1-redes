package pacman

import (
	"math/rand/v2"

	rawsockets "pacman-redes/lib/rawSockets"
)

type Direction int

const (
	Up Direction = iota
	Right
	Down
	Left
)

func (d Direction) turnLeft() Direction {
	return (d + 3) % 4
}

func (d Direction) turnRight() Direction {
	return (d + 1) % 4
}

func (d Direction) dx() int {
	switch d {
	case Right:
		return 1
	case Left:
		return -1
	default:
		return 0
	}
}

func (d Direction) dy() int {
	switch d {
	case Down:
		return 1
	case Up:
		return -1
	default:
		return 0
	}
}

func (p *Position) detectCollision(p2 *Position) bool {
	return p.x == p2.x && p.y == p2.y
}

func (m *Map) isWall(x, y int) bool {
	if x < 0 || y < 0 || x >= 40 || y >= 40 {
		return true
	}
	return m.grid[y][x] == 'X'
}

func (m *Map) moveGhostInDir(g *Ghost, dir Direction) bool {
	nx := int(g.ent.pos.x) + dir.dx()
	ny := int(g.ent.pos.y) + dir.dy()
	if m.isWall(nx, ny) {
		return false
	}
	g.ent.pos.x = uint8(nx)
	g.ent.pos.y = uint8(ny)
	return true
}

// Red: regra da mão esquerda — segue reto; ao colidir, vira à esquerda
func (m *Map) moveRed(g *Ghost, dir *Direction) {
	if !m.moveGhostInDir(g, *dir) {
		*dir = dir.turnLeft()
	}
}

// Blue: regra da mão direita — segue reto; ao colidir, vira à direita
func (m *Map) moveBlue(g *Ghost, dir *Direction) {
	if !m.moveGhostInDir(g, *dir) {
		*dir = dir.turnRight()
	}
}

// Green: alterna direita e esquerda a cada colisão
func (m *Map) moveGreen(g *Ghost, dir *Direction, turnRight *bool) {
	if !m.moveGhostInDir(g, *dir) {
		if *turnRight {
			*dir = dir.turnRight()
		} else {
			*dir = dir.turnLeft()
		}
		*turnRight = !*turnRight
	}
}

// Yellow: aleatório ao colidir
func (m *Map) moveYellow(g *Ghost, dir *Direction) {
	if !m.moveGhostInDir(g, *dir) {
		*dir = Direction(rand.IntN(4))
	}
}

func (gs *GameState) moveGhosts() {
	for i := range gs.GameMap.ghosts {
		g := &gs.GameMap.ghosts[i]
		switch g.ent.symbol {
		case 'R':
			gs.GameMap.moveRed(g, &g.dir)
		case 'B':
			gs.GameMap.moveBlue(g, &g.dir)
		case 'G':
			gs.GameMap.moveGreen(g, &g.dir, &g.turnRight)
		case 'Y':
			gs.GameMap.moveYellow(g, &g.dir)
		}
	}
}

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
	gs.moveGhosts()

	// detectar colisao

	return nil
}
