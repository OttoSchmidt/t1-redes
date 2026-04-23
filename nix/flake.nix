{
  description = "ISO Pacman - NixOS 25.11";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-25.11";
  };

  outputs = { self, nixpkgs }: {
    nixosConfigurations.live-iso = nixpkgs.lib.nixosSystem {
      system = "x86_64-linux";
      modules = [
        # Base da ISO oficial
        "${nixpkgs}/nixos/modules/installer/cd-dvd/installation-cd-minimal.nix"
        
        # Importa o seu arquivo de configuração
        ./configuration.nix

        # Ajustes específicos para o ambiente Live
        ({ pkgs, ... }: {
          networking.hostName = "pacman";

          # não precisa compactar com força máxima se estiver testando
          isoImage.squashfsCompression = "gzip -Xcompression-level 1";
          
          # Garante que o suporte a Wi-Fi e redes comuns esteja ativo na ISO
          networking.networkmanager.enable = true;
        })
      ];
    };
  };
}
