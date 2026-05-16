package pacman

import (
	"fmt"
	"math/rand/v2"
	"os"
	"strings"
)

type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

func Abs[T Signed](x T) T {
	if x < 0 {
		return -x
	}
	return x
}

// inserir entidade numa posicao aleatoria no mapa
func randomPosition(ent Entity, m Map) {
	for {
		line := rand.IntN(40)
		column := rand.IntN(40)

		if m.grid[line][column] != "X" {
			ent.setPos(line, column)
			break
		}
	}
}

// inserir entidades num grid seguindo a ordem: pastilhas -> fantasmas -> pacman
func (m *Grid) insertEntities(pacman Pacman, ghosts []Ghost, coins []Coin) {
	for _, e := range coins {
		(*m)[e.ent.pos.y][e.ent.pos.x] = e.ent.symbol
	}
	for _, e := range ghosts {
		(*m)[e.ent.pos.y][e.ent.pos.x] = e.ent.symbol
	}
	(*m)[pacman.ent.pos.y][pacman.ent.pos.x] = pacman.ent.symbol
}

// retornar mapa considerando o raio do pacman
func getMapWindow(g Grid, pacman Pacman, windowSize int) Grid {
	sl := max(int(pacman.ent.pos.y) - windowSize, 0)
	el := min(int(pacman.ent.pos.y) + windowSize, 39)
	sc := max(int(pacman.ent.pos.x) - windowSize, 0)
	ec := min(int(pacman.ent.pos.x) + windowSize, 39)

	sub := g[sl:(el+1)]
	for i := range sub {
		sub[i] = sub[i][sc:(ec+1)]
	}

	return sub
}

func (g *Grid) ToString(center Position) string {
	windowSize := len(*g)/2
	var gridComplete strings.Builder

	for i := 0; i < 84; i++ { // topo do frame
		gridComplete.WriteString("█")
	}
	gridComplete.WriteString("\n")

	for i := int(center.y) - windowSize; i > 0; i-- { // vazio superior
		gridComplete.WriteString(fmt.Sprintf("██%s", BLACK)) // lado esquerdo frame
		for i := 0; i < 80; i++ { // fundo do mapa
			gridComplete.WriteString("█")
		}
		gridComplete.WriteString(fmt.Sprintf("%s██\n", NC)) // lado direito frame
	}

	for _, line := range *g { // centro
		gridComplete.WriteString("██") // frame esquerdo
		for i := int(center.x) - windowSize; i > 0; i-- { // vazio esquerdo
			gridComplete.WriteString(fmt.Sprintf("%s██%s", BLACK, NC))
		}
		for _, c := range line { // conteudo mapa
			gridComplete.WriteString(c)
		}
		for i := int(center.x) + windowSize; i < 40; i++ { // vazio direito
			gridComplete.WriteString(fmt.Sprintf("%s██%s", BLACK, NC))
		}
		gridComplete.WriteString("██\n") // frame direito
	}

	for i := int(center.y) + windowSize; i < 40; i++ { // vazio inferior
		gridComplete.WriteString(fmt.Sprintf("██%s", BLACK)) // lado esquerdo frame
		for i := 0; i < 80; i++ { // fundo mapa
			gridComplete.WriteString("█")
		}
		gridComplete.WriteString(fmt.Sprintf("%s██\n", NC)) // lado direito frame
	}

	for i := 0; i < 84; i++ { // baixo frame
		gridComplete.WriteString("█")
	}
	gridComplete.WriteString("\n")

	return gridComplete.String()
}

// criar visualizacao do mapa com as entidades
func (m *Map) ToString() string {
	// copiar mapa
	filledMap := make(Grid, len((*m).grid))
	for i := range filledMap {
		filledMap[i] = make([]string, len((*m).grid[i]))
		copy(filledMap[i], (*m).grid[i])
	}

	filledMap.insertEntities(m.pacman, m.ghosts, m.coins)
	windowedMap := getMapWindow(filledMap, m.pacman, m.windowSize)

	return windowedMap.ToString(m.pacman.ent.pos)
}

// converter visualizacao do mapa com entidades para vetor de bytes
// ira incluir a posicao do pacman e quantidade de linhas
func (m *Map) ToBytes() []byte {
	// copiar mapa
	filledMap := make(Grid, len((*m).grid))
	for i := range filledMap {
		filledMap[i] = make([]string, len((*m).grid[i]))
		copy(filledMap[i], (*m).grid[i])
	}

	filledMap.insertEntities(m.pacman, m.ghosts, m.coins)
	windowedMap := getMapWindow(filledMap, m.pacman, m.windowSize)

	// converter matriz para vetor
	var map1D []byte
	map1D = append(map1D, uint8(len(windowedMap)), uint8(m.pacman.ent.pos.x), uint8(m.pacman.ent.pos.y))
	for _, l := range windowedMap {
		for _, c := range l {
			map1D = append(map1D, []byte(c)...)
			map1D = append(map1D, 0x0)
		}
	}

	return map1D
}

func GridFromBytes(stream []byte) (Grid, Position) {
	numLines := uint8(stream[0])
	center := Position{
		x: uint8(stream[1]),
		y: uint8(stream[2]),
	}

	map1D := stream[3:]

	// extrair celulas do mapa (substrings)
	substringStart := 0
	substringSize := 0
	cells := make([]string, 0)
	for i := 0; i < len(map1D); i++ {
		if map1D[i] != 0x0 {
			substringSize++
		} else {
			cell := map1D[substringStart:substringStart+substringSize]
			cells = append(cells, string(cell))

			substringStart = i+1
			substringSize = 0
		}
	}

	grid := make(Grid, numLines)
	numColumns := len(cells)/int(numLines)
	for i := range grid {
		grid[i] = make([]string, numColumns)
		copy(grid[i], cells[i*numColumns:(i+1)*numColumns+1])
	}

	return grid, center
}

// ler mapa e entidades de um csv
func (s *GameState) ReadMapCsv(csv string) error {
	file, err := os.ReadFile(csv)
	if err != nil {
		return fmt.Errorf("erro ao abrir mapa: %w\n", err)
	}

	s.GameMap.grid = make(Grid, 40)
	s.GameMap.ghosts = make([]Ghost, 0)
	s.GameMap.coins = make([]Coin, 0)
	s.GameMap.windowSize = 1

	lines := strings.Split(string(file), "\n")
	for i := 0; i < 40; i++ {
		s.GameMap.grid[i] = make([]string, 40)
		cells := strings.Split(string(lines[i]), ";")
		if len(cells) != 40 {
			return fmt.Errorf(".csv com quant. colunas erradas (atual: %d | esperado: 40) na linha %d\n", len(cells), i+1)
		}

		for j, c := range cells {
			switch c {
			case "X", "x":
				s.GameMap.grid[i][j] = "██"
			case "0":
				s.GameMap.grid[i][j] = fmt.Sprintf("%s██%s", BLACK, NC)
			case "P", "p":
				p, err := createPacman(j, i)
				if err != nil {
					return err
				}
				s.GameMap.pacman = p
			case "R", "r", "B", "b", "Y", "y", "G", "g":
				newGhost, err := createGhost(j, i, byte(c[0]))
				if err != nil {
					return err
				}
				s.GameMap.ghosts = append(s.GameMap.ghosts, newGhost)
			case "1", "2", "3", "4", "5", "6":
				newCoin, err := createCoin(j, i, byte(c[0]))
				if err != nil {
					return err
				}
				s.GameMap.coins = append(s.GameMap.coins, newCoin)
			case "":
				continue
			default:
				return fmt.Errorf("simbolo em .csv nao reconhecido: %s (%d)\n", c, c[0])
			}			
		}
	}

	return nil
}
