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

		if m.grid[line][column] != 'X' {
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
func getMapWindow(g Grid, pacman Pacman, windowSize uint8) Grid {
	sl := max(int(pacman.ent.pos.y) - int(windowSize), 0)
	el := min(int(pacman.ent.pos.y) + int(windowSize), 39)
	sc := max(int(pacman.ent.pos.x) - int(windowSize), 0)
	ec := min(int(pacman.ent.pos.x) + int(windowSize), 39)

	sub := g[sl:(el+1)]
	for i := range sub {
		sub[i] = sub[i][sc:(ec+1)]
	}

	return sub
}

func (g *Grid) ToString(center Position, radius uint8) string {
	var gridComplete strings.Builder

	for i := 0; i < 84; i++ { // topo do frame
		gridComplete.WriteString("█")
	}
	gridComplete.WriteString("\n")

	for i := int(center.y) - int(radius); i > 0; i-- { // vazio superior
		gridComplete.WriteString(fmt.Sprintf("██%s", BLACK)) // lado esquerdo frame
		for i := 0; i < 80; i++ { // fundo do mapa
			gridComplete.WriteString("█")
		}
		gridComplete.WriteString(fmt.Sprintf("%s██\n", NC)) // lado direito frame
	}

	for _, line := range *g { // centro
		gridComplete.WriteString(fmt.Sprintf("██%s", BLACK)) // frame esquerdo
		for i := int(center.x) - int(radius); i > 0; i-- { // vazio esquerdo
			gridComplete.WriteString("██")
		}
		gridComplete.WriteString(NC)

		for _, c := range line { // conteudo mapa
			var element string
			switch c {
			case ' ':
				element = fmt.Sprintf("%s██%s", BLACK, NC)
			case 'X':
				element = "██"
			case 'P':
				element = fmt.Sprintf("%s P%s", YELLOW, NC)
			case 'C':
				element = fmt.Sprintf("%s C%s", YELLOW, NC)
			case 'Y':
				element = fmt.Sprintf("%s Y%s", YELLOW, NC)
			case 'R':
				element = fmt.Sprintf("%s R%s", RED, NC)
			case 'G':
				element = fmt.Sprintf("%s G%s", GREEN, NC)
			case 'B':
				element = fmt.Sprintf("%s B%s", BLUE, NC)
			}

			gridComplete.WriteString(element)
		}

		gridComplete.WriteString(BLACK)
		for i := int(center.x) + int(radius) + 1; i < 40; i++ { // vazio direito
			gridComplete.WriteString("██")
		}
		gridComplete.WriteString(fmt.Sprintf("%s██\n", NC)) // frame direito
	}

	for i := int(center.y) + int(radius) + 1; i < 40; i++ { // vazio inferior
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
		filledMap[i] = make([]byte, len((*m).grid[i]))
		copy(filledMap[i], (*m).grid[i])
	}

	filledMap.insertEntities(m.pacman, m.ghosts, m.coins)
	windowedMap := getMapWindow(filledMap, m.pacman, m.windowSize)

	return windowedMap.ToString(m.pacman.ent.pos, m.windowSize)
}

// converter visualizacao do mapa com entidades para vetor de bytes
// ira incluir a posicao do pacman e quantidade de linhas
func (m *Map) ToBytes() []byte {
	// copiar mapa
	filledMap := make(Grid, len((*m).grid))
	for i := range filledMap {
		filledMap[i] = make([]byte, len((*m).grid[i]))
		copy(filledMap[i], (*m).grid[i])
	}

	filledMap.insertEntities(m.pacman, m.ghosts, m.coins)
	windowedMap := getMapWindow(filledMap, m.pacman, m.windowSize)

	// converter matriz para vetor
	var map1D []byte
	map1D = append(map1D, uint8(len(windowedMap)), uint8(m.pacman.ent.pos.x), uint8(m.pacman.ent.pos.y), uint8(m.windowSize))
	for _, l := range windowedMap {
		map1D = append(map1D, l...)
	}

	return map1D
}

func GridFromBytes(stream []byte) (Grid, Position, uint8) {
	if len(stream) < 4 {
		return nil, Position{}, 0
	}

	map1D := stream[4:]
	numLines := uint8(stream[0])
	numColumns := (len(map1D))/int(numLines)
	center := Position{
		x: uint8(stream[1]),
		y: uint8(stream[2]),
	}

	grid := make(Grid, numLines)
	for i := range grid {
		grid[i] = make([]byte, numColumns)
		copy(grid[i], map1D[i*numColumns:(i+1)*numColumns+1])
	}

	return grid, center, stream[3]
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
		s.GameMap.grid[i] = make([]byte, 40)
		cells := strings.Split(string(lines[i]), ";")
		if len(cells) != 40 {
			return fmt.Errorf(".csv com quant. colunas erradas (atual: %d | esperado: 40) na linha %d\n", len(cells), i+1)
		}

		for j, c := range cells {
			switch c {
			case "X", "x":
				s.GameMap.grid[i][j] = 'X'
			case "0":
				s.GameMap.grid[i][j] = ' '
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
