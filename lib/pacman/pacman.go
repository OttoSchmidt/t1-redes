package pacman

import "fmt"

const RED = "\033[31m"
const GREEN = "\033[32m"
const YELLOW = "\033[33m"
const BLUE = "\033[34m"
const BLACK = "\033[30m"
const NC = "\033[0m"

type Grid [][]string

type Map struct {
	grid Grid
	windowSize int

	pacman Pacman
	ghosts []Ghost
	coins []Coin
}

type GameState struct {
	GameMap Map
	round int
}

func (gs *GameState) IncrementRound() {
	gs.round++
	if gs.round % 5 == 0 {
		gs.GameMap.windowSize++
	}
}

type Position struct {
	x, y uint8
}

type Entity struct {
	pos Position
	symbol string	
}

func (e *Entity) setPos(x, y int) {
	e.pos.x = uint8(x)
	e.pos.y = uint8(y)
}

type Pacman struct {
	ent Entity
}

func createPacman(x, y int) (Pacman, error) {
	pacman := Pacman{
		ent: Entity{
			symbol: fmt.Sprintf("%s P%s", YELLOW, NC),
		},
	}
	pacman.ent.setPos(x, y)

	return pacman, nil
}

type Ghost struct {
	ent Entity
}

func createGhost(x, y int, s byte) (Ghost, error) {
	var color string
	switch s {
	case 'Y':
		color = YELLOW
	case 'B':
		color = BLUE
	case 'R':
		color = RED
	case 'G':
		color = GREEN
	}

	g := Ghost {
		ent: Entity{
			symbol: fmt.Sprintf("%s %s%s", color, string(s), NC),
		},
	}
	g.ent.setPos(x, y)

	return g, nil
}

type Coin struct {
	ent Entity
	fileName string		
}

func createCoin(x, y int, id byte) (Coin, error) {
	var ext string
	switch id {
	case '1', '2':
		ext = "txt"
	case '3', '4':
		ext = "jpg"
	case '5', '6':
		ext = "mp4"
	default:
		return Coin{}, fmt.Errorf("erro ao criar pastilha: id nao reconhecido => %c\n", id)
	}

	file := fmt.Sprintf("./files/%s.%s", string(id), ext)
	coin := Coin {
		ent: Entity{
			symbol: fmt.Sprintf("%s C%s", YELLOW, NC),
		},
		fileName: file,
	}
	coin.ent.setPos(x, y)

	return coin, nil
} 
