package pacman

type GameState struct {
	round int
}

type Position struct {
	x, y uint16
}

type Entity struct {
	pos Position
	symbol byte	
}

func (e *Entity) setPos(x, y int) {
	e.pos.x = uint16(x)
	e.pos.y = uint16(y)
}

func (e *Entity) detectCollision(e2 *Entity) bool {
	return false
}


type Pacman struct {
	ent Entity
	mapWindow int
}

type Ghost struct {
	ent Entity
	typ byte
}

func (g *Ghost) movement() {

}

type Coin struct {
	ent Entity
	fileName string		
}

