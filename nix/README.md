## Construir ISO

> Necessário ter o Nix instalado.

Para construir a ISO, basta rodar o comando:

```bash
sudo nix build ".#nixosConfigurations.live-iso.config.system.build.isoImage"
```

Isso irá gerar a ISO em `./result/iso/nixos-*.iso`. Essa ISO pode ser utilizada numa VM ou num ambiente de live boot num computador físico.

## Rodando o sistema

### VM

> Recomendo utilizar o virt-manager para criar a VM.

#### Montando diretório do projeto

Com o filesystem do projeto configurado com o nome `pacman`, basta rodar o comando:

```bash
mkdir ~/pacman
sudo mount -t virtiofs pacman ~/pacman
```

O diretório do projeto estará disponível em `~/pacman`. Lembre-se de desmontar o diretório após o uso:

```bash
sudo umount ~/pacman
```

#### Instalação

Para instalar o sistema na VM de BIOS/Legacy, precisamos formatar o disco virtual (nesse caso, o `/dev/vda`) e montar a partição.
Basta rodar os seguintes comandos:

```bash
# cria a tabela de partições do tipo MBR (msdos)
sudo parted /dev/vda -- mklabel msdos

# cria uma particao primaria pegando todo o espaço do disco
sudo parted /dev/vda -- mkpart primary 1MiB 100%

# ativa a flag de boot nessa particao
sudo parted /dev/vda -- set 1 boot on

# formatar a particao criada
sudo mkfs.ext4 -L nixos /dev/vda1

# montar a particao
sudo mount /dev/vda1 /mnt
```

Depois, precisamos gerar o `hardware-configuration.nix` para a VM. Para isso, basta rodar o comando:

```bash
sudo nixos-generate-config --root /mnt
cp /mnt/etc/nixos/hardware-configuration.nix ~/pacman/nix
```

E então, dentro de `~/pacman/nix`, execute o seguinte comando para instalar o SO

```bash
sudo nixos-install --flake .#vm-persistente
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
