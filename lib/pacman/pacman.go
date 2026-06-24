package pacman

import "fmt"

const BACKGROUND = "\033[48;5;16m"
const WALL = "\033[38;5;17m"
const RED = "\033[31m"
const GREEN = "\033[32m"
const YELLOW = "\033[33m"
const BLUE = "\033[34m"
const BLACK = "\033[38;5;16m"
const NC = "\033[0m"

type Grid [][]byte

type Map struct {
	grid       Grid
	windowSize uint8

	pacman Pacman
	ghosts []Ghost
	coins  []Coin
}

type GameState struct {
	GameMap        Map
	round  		   int
	CoinsCollected int
}

func (gs *GameState) IncrementRound() {
	gs.round++
	if gs.round%5 == 0 && gs.GameMap.windowSize < 255 {
		gs.GameMap.windowSize++
	}
}

type Position struct {
	x, y uint8
}

type Entity struct {
	pos    Position
	symbol byte
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
			symbol: 'P',
		},
	}
	pacman.ent.setPos(x, y)

	return pacman, nil
}

type Ghost struct {
	ent       Entity
	dir       Direction
	turnRight bool
}

func createGhost(x, y int, s byte) (Ghost, error) {
	g := Ghost{
		ent: Entity{
			symbol: s,
		},
	}
	g.ent.setPos(x, y)

	return g, nil
}

type Coin struct {
	ent      Entity
	id 		 byte
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
	coin := Coin{
		ent: Entity{
			symbol: 'C',
		},
		fileName: file,
	}
	coin.ent.setPos(x, y)
	coin.id = id

	return coin, nil
}
