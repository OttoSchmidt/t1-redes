## Decisões

- Timeout inicial é 500ms. Considerando que as duas máquinas estarão conectadas entre si, é um número bem generoso;
- A cada tentativa, o tempo de timeout dobra (máx: 4s), pois é bem provável que a conexão caiu, então não faz sentido continuar com um timeout pequeno;
- O número de tentativas é 50, pois é um número grande o suficiente para tentar se reconectar sem perder o status do jogo;
- A cada transmissão, o conteúdo é dividido em mensagens de no máximo 32 bytes. E depois do conteúdo transmitido, é enviado uma mensagem de tipo 16 para sinalizar o fim da transmissão.
