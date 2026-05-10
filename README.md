## Decisões

- Pré-calculamos o CRC para cada byte possível (256 valores) para otimizar o tempo p/ transmissao. Então, para calcular o CRC de um conteúdo, basta iterar sobre cada byte do conteúdo e fazer uma operação XOR com o valor pré-calculado do CRC para aquele byte [(fonte)](https://web.archive.org/web/20230525024916/http://sbs-forum.org/marcom/dc2/20_crc-8_firmware_implementations.pdf).
- Timeout inicial é 500ms. Considerando que as duas máquinas estarão conectadas entre si, é um número bem generoso;
- A cada tentativa, o tempo de timeout dobra (máx: 4s), pois é bem provável que a conexão caiu, então não faz sentido continuar com um timeout pequeno;
- O número de tentativas é 50, pois é um número grande o suficiente para tentar se reconectar sem perder o status do jogo;
- A cada transmissão, o conteúdo é dividido em mensagens de no máximo 32 bytes. E depois do conteúdo transmitido, é enviado uma mensagem de tipo 16 para sinalizar o fim da transmissão.
- Caso o byte na posição 13 da mensagem seja 0x88 ou 0x81, o byte na posição 14 será 0xff. Isso é necessário para evitar que a mensagem seja interpretada como um comando de VLAN, já que o byte 0x88 é o Ethertype para VLAN e o byte 0x81 é o Ethertype para QinQ. Caso o byte seja adicionado, ele não afeta o campo tamanho da mensagem, pois o byte 0xff é um valor de preenchimento que não é considerado parte do conteúdo da mensagem.
