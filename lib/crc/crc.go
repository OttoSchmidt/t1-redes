package crc

import (
	"fmt"
	debug "pacman-redes/lib/debug"
)

// fonte: https://web.archive.org/web/20230525024916/http://sbs-forum.org/marcom/dc2/20_crc-8_firmware_implementations.pdf

type lookupTable [16][16]byte

func (t *lookupTable) generate() {
	for i := 0; i < 16; i++ {
		for j := 0; j < 16; j++ {
			// valor inicial de crc é o XOR do índice i e j, que representa 
			// os 4 bits superiores e inferiores de um byte
			crc := byte((i << 4) | j)

			// calcular o crc realizando 8 iteracoes de deslocamento bit a bit,
			// aplicando o polinômio quando o bit mais significativo for 1
			for k := 0; k < 8; k++ {
				if (crc & 0x80) != 0 {
					crc = (crc << 1) ^ crcPolynomial
				} else {
					crc <<= 1
				}
			}

			crcTable[i][j] = crc
		}
	}
}

func (t *lookupTable) print() {
	for i := 0; i < 16; i++ {
		for j := 0; j < 16; j++ {
			fmt.Printf("%02X ", t[i][j])
		}
		fmt.Println()
	}
}

var crcTable lookupTable
const crcPolynomial = 0x07

func init() {
	debug.PrintLog("Gerando tabela de CRC-8 com polinômio 0x%02X.\n", crcPolynomial)

	crcTable.generate()

	debug.PrintLog("Tabela de CRC-8 gerada:\n")
	if debug.Debug {
		crcTable.print()
	}
}

// Calcula o CRC-8 para o buffer fornecido, o qual deve incluir somente 
// os campos de tamanho, ID, tipo e dados
func CalculateCRC(data []byte) byte {
	debug.PrintLog("Calculando CRC para os dados: %x\n", data)
	
	crc := byte(0)
	for _, b := range data {
		value := b ^ crc // XOR do byte atual com o CRC acumulado
		crc = crcTable[value>>4][value&0x0F] // usar os 4 bits superiores e inferiores para indexar a tabela
	}

	debug.PrintLog("CRC calculado: %02X\n", crc)

	return crc
}

// Verifica se o CRC recebido é válido para os dados fornecidos. O buffer deve incluir
// somente os campos de tamanho, ID, tipo e dados.
func VerifyCRC(data []byte, receivedCRC byte) bool {
	debug.PrintLog("CRC recebido: %02X\n", receivedCRC)
	calculatedCRC := CalculateCRC(data)

	return receivedCRC == calculatedCRC
}