package pacman

import (
	"rand"
)

type Map [40][40]byte

func (m *Map) toBytes(pacman Pacman, ent []Entity) []byte {
	windowedMap := getMapWindow(*m, pacman)

	// inserir entidades no mapa
	// se estiverem perto do pacman

	return nil
}

func randomPosition(ent Entity, m Map) {
	for {
		line := rand.Int() % 40
		column := rand.Int() % 40

		if m[line][column] != byte('X') {
			ent.setPos(line, column)
			break
		}
	}
}

func readMapCsv(csv string) Map {
	return Map{}
}

// considerar borda do mapa
// fazer uma copia do slice
func getMapWindow(m Map, pacman Pacman) [][]byte {
	sl := pacman.ent.pos.y - 

	var newMap [][]byte
	copy(newMap, [][]byte(m[sl:el][sc:ec]))

	return nil
}

