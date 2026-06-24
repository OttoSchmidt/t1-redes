# Relatório — Trabalho T1: PacMan Remoto

## Visão Geral

Implementação de um jogo Pac-Man "no escuro" na modalidade cliente-servidor, com comunicação exclusiva via **RAW Sockets** (AF_PACKET, camada 2 Ethernet). Todo o protocolo de enlace foi construído do zero, sem uso de TCP/IP.

---

## Protocolo de Comunicação

### Enquadramento

Cada mensagem é encapsulada em um frame binário customizado:

```
[0x7E] [tamanho(5b) + seq_hi(3b)] [seq_lo(3b) + tipo(5b)] [0–31B dados] [CRC-8] [padding ≥ 15B]
```

- **Marcador de início:** `0x7E`
- **Número de sequência:** 6 bits (valores 0–63), compartilhado entre os lados e incrementado a cada mensagem enviada e recebida
- **Tipo de pacote:** 5 bits — ACK, NACK, Visualize, Init, Data, TxtFile, JpgFile, Mp4File, MoveUp/Down/Left/Right, End, EndConn, Error
- **Dados:** até 31 bytes por frame
- **Proteção VLAN:** se o byte na posição 9 do payload for `0x88` ou `0x81`, um byte `0xFF` é inserido na posição 10 para evitar que switches interpretem o frame como VLAN-tagged

### Detecção de Erros

CRC-8 com polinômio `0x07`, calculado sobre os bytes do frame excluindo o marcador de início. Calculado via lookup table inicializada em tempo de compilação.

### Controle de Fluxo — Para-e-Espera

Implementado em `SendMessage`: o remetente envia um frame e aguarda ACK antes de enviar o próximo. Em caso de NACK, retransmite imediatamente com reset do contador de tentativas. Em caso de timeout, retransmite com **backoff exponencial**: timeout inicial de 500ms, dobrado a cada tentativa, limitado a 4000ms, com máximo de 50 tentativas. ACK e NACK são enviados sem aguardar resposta.

Detecção de duplicatas: se o número de sequência recebido for igual ao último recebido com sucesso, o pacote é tratado como retransmissão e o último ACK é reenviado.

### Transferência de Arquivos

Conteúdo longo (visualizações e arquivos) é fragmentado em chunks de 31 bytes, cada um enviado com `SendMessage`. Um pacote `End` sinaliza o fim da transmissão. O receptor reconstrói o conteúdo e, no caso de arquivos, verifica espaço em disco antes de aceitar, salva em `/tmp` e abre com `xdg-open`.

---

## Jogo

### Labirinto e Entidades

O mapa 40×40 é lido de um arquivo CSV com separador `;`. Símbolos: `X` (parede), `0` (vazio), `P` (Pac-Man), `R`/`B`/`G`/`Y` (fantasmas), `1`–`6` (pastilhas). Entidades ausentes no CSV são posicionadas aleatoriamente em células livres, sem sobreposição.

### Visibilidade

O cliente só exibe a janela ao redor do Pac-Man. A cada 5 rodadas a visibilidade aumenta em 1 célula (`windowSize++`). O servidor serializa apenas a subgrade visível e a envia ao cliente.

### Movimentos dos Fantasmas

| Fantasma | Comportamento |
|---|---|
| Vermelho | Regra da mão esquerda — segue reto, vira à esquerda ao colidir com parede |
| Azul | Regra da mão direita — segue reto, vira à direita ao colidir com parede |
| Verde | Alterna entre virar à esquerda e à direita a cada colisão |
| Amarelo | Escolhe direção aleatória ao colidir com parede |

### Colisões e Arquivos

- **Pastilha coletada:** o servidor abre o arquivo correspondente (`1.txt`–`6.mp4`) e o transmite ao cliente via `SendFile`, que o salva em `/tmp` e o abre automaticamente.
- **Encontro com fantasma:** envia `jumpscare.mp4` ao cliente.

### Comunicação por Rodada

1. **Cliente** envia um pacote de movimento (tamanho fixo)
2. **Servidor** atualiza posição do Pac-Man, move fantasmas, detecta colisões
3. **Servidor** serializa a janela visível e envia ao cliente (tamanho variável, cresce a cada 5 rodadas)
4. Se houver pastilha ou fantasma, envia o arquivo correspondente antes da visualização

---

## Decisões de Implementação

- **Linguagem:** Go, com uso da biblioteca Bubbletea para o TUI do cliente
- **Movimentos dos fantasmas:** adotamos os comportamentos descritos na Seção "Jogo" do enunciado (p.2), onde Verde alterna direita/esquerda e Amarelo é aleatório
- **Log separado:** implementado via named pipe (`/tmp/pacman_pipe`) lido por um terminal ptyxis separado, exibindo todas as mensagens enviadas e recebidas em tempo real
- **Ambiente reproduzível:** NixOS com flake para ISO bootável e VM, garantindo o mesmo ambiente de execução em qualquer máquina
