## Construir ISO

> Necessário ter o Nix instalado.

Para construir a ISO, basta rodar o comando:

```bash
sudo nix build ".#nixosConfigurations.live-iso.config.system.build.isoImage"
```

Isso irá gerar a ISO em `./result/iso/nixos-*.iso`. Essa ISO pode ser utilizada numa VM ou num ambiente de live boot num computador físico.

## Montar Diretório do Projeto

### VM

> Recomendo utilizar o virt-manager para criar a VM.

Com o filesystem do projeto configurado com o nome `pacman`, basta rodar o comando:

```bash
mkdir ~/pacman
sudo mount -t virtiofs pacman ~/pacman
```

O diretório do projeto estará disponível em `~/pacman`. Lembre-se de desmontar o diretório após o uso:

```bash
sudo umount ~/pacman
```

### Live Boot

No ambiente de live boot, teremos que montar o diretório do projeto a partir do SSH. Para isso, precisamos primeiro habilitar o SSH no computador com o projeto. Para isso, basta rodar o comando:

```bash
sudo systemctl start sshd
```

Em seguida, podemos montar o diretório do projeto utilizando o SSHFS. Para isso, basta rodar o comando:

```bash
sshfs usuario_do_laptop@ip_do_laptop:/caminho/do/projeto ~/pacman
```

O diretório do projeto estará disponível em `~/pacman`. Lembre-se de desmontar o diretório após o uso:

```bash
fusermount -u ~/pacman
```
