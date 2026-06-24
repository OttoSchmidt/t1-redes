# Pacman - Trabalho de Redes

Este projeto é uma implementação do jogo Pacman comunicando-se através de _Raw Sockets_.

## Decisões de Implementação e Protocolo

- **Otimização de CRC:** Pré-calculamos o CRC para cada byte possível (256 valores) para otimizar o tempo para transmissão. Então, para calcular o CRC de um conteúdo, basta iterar sobre cada byte do conteúdo e fazer uma operação XOR com o valor pré-calculado do CRC para aquele byte [(fonte)](https://web.archive.org/web/20230525024916/http://sbs-forum.org/marcom/dc2/20_crc-8_firmware_implementations.pdf).
- **Tratamento de Timeout e Controle de Duplicatas:** Timeout inicial é 500ms. Considerando que as duas máquinas estarão conectadas diretamente entre si, é um valor bem generoso. A cada tentativa falha, o tempo de timeout dobra (limitado a um máximo de 4s). Isso ocorre pois é provável que a conexão tenha caído, então não faz sentido sobrecarregar a rede com tentativas seguidas de reconexão. Também conta com detecção de duplicatas, reenviando o último ACK em caso de pacotes repetidos.
- **Tamanho Mínimo de Mensagem:** Caso uma mensagem tenha menos que 14 bytes, são inseridos bytes nulos no final da mensagem para garantir o envio correto e alcançar os limites exigidos na rede.
- **Limites de Retransmissão:** O número máximo de tentativas de envio de pacotes é de 50. É um número grande o suficiente para tentar se reconectar sem perder o status atual do jogo.
- **Transmissão de Arquivos e Segmentação:** A cada transmissão, o conteúdo pesado (como imagens e vídeos) é segmentado e dividido em mensagens contendo no máximo 31 bytes de payload. Após todo o conteúdo ser transmitido com sucesso, é enviada uma mensagem especial indicando o fim da transmissão. O cliente então reconstrói o artefato recebido de forma assíncrona, salva em `/tmp` e o abre automaticamente utilizando `xdg-open`.
- **Prevenção de Colisão com Ethertype (VLAN):** Caso o byte na posição 13 da mensagem (onde normalmente ficaria o Ethertype) seja `0x88` ou `0x81`, o byte na posição 14 será definido como `0xff`. Isso previne que roteadores/switches interpretem erroneamente o pacote como uma tag de VLAN ou QinQ. O byte de padding `0xff` não afeta o tamanho real interpretado do payload e é descartado no recebimento.

## Detalhes da Regra de Negócio do Jogo

### Visibilidade Dinâmica "No Escuro"

- O cliente opera sem conhecimento do mapa inteiro, exibindo inicialmente apenas uma pequena janela ao redor do Pac-Man. A cada 5 rodadas, a visibilidade aumenta em 1 célula. O servidor é responsável por isolar apenas a subgrade visível e enviá-la serializada ao cliente a cada rodada.

### Colisões e Premiações

- **Pastilhas Especiais:** Ao coletar uma pastilha (marcadas de 1 a 6 no mapa), o servidor abre o arquivo de mídia correspondente (ex: `1.txt` a `6.mp4`) e inicia a transferência para o cliente.
- **Fantasmas:** O encontro com um fantasma dispara o envio de um arquivo de susto (`jumpscare.mp4`) que é reproduzido pelo cliente.

### Mapa e Posições (CSV)

- Para evitar que o pacman, fantasmas e moedas sejam gerados em posições inacessíveis do cenário (por exemplo, dentro das letras P e R do mapa da UFPR), é possível inserir o caractere `-` no arquivo `.csv` do mapa. Este caractere impede a geração de entidades naquela posição, porém fica invisível visualmente para os jogadores.

### Comportamento dos Fantasmas

Os fantasmas possuem lógicas de movimentação distintas:

- **Vermelho (`R`):** Segue a regra da mão esquerda. Ele se move em linha reta e, ao colidir com uma parede, vira sempre à esquerda.
- **Azul (`B`):** Segue a regra da mão direita. Ele se move em linha reta e, ao colidir com uma parede, vira sempre à direita.
- **Verde (`G`):** Alterna o comportamento. A cada colisão com parede, ele intercala sua decisão entre virar à direita e virar à esquerda.
- **Amarelo (`Y`):** Movimentação aleatória. Ao bater numa parede, ele escolhe aleatoriamente uma nova direção possível.

## Arquitetura de Interface e Logs

### Janela de Log Desacoplada

Como a interface do cliente e servidor operam direto no terminal (usando a biblioteca gráfica _Bubbletea_), as saídas de logs (_debugs_, informações de pacotes) poluiriam a interface e impediriam de renderizar o mapa do jogo. A solução:

- O sistema cria um canal do Go (`chan string`) e um _Named Pipe_ (`/tmp/pacman_pipe`).
- Uma _goroutine_ consome os eventos que chegam no canal e escreve no Pipe em background.
- É disparado um comando assíncrono para abrir um novo terminal nativo (`ptyxis --new-window`), que fará exclusivamente a leitura dos logs usando `cat < /tmp/pacman_pipe`.
- **Elevação de Privilégios (Root):** Como o jogo exige criação de _Raw Sockets_, ele deve ser executado pelo usuário `root`. Porém, o `root` muitas vezes não tem permissão para abrir janelas do sistema gráfico diretamente. Para solucionar isso, o comando que abre o terminal de log é injetado com o usuário original (`SUDO_USER`) e recria a variável `XDG_RUNTIME_DIR` para conseguir "voltar" pro usuário normal e abrir a janela gráfica de _logs_ corretamente na sessão em andamento.
