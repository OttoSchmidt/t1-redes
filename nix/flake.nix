{
  description = "Pacman OS - ISO e VM";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-25.11";
  };

  outputs = { self, nixpkgs }: {
    nixosConfigurations = {
      
      # Live ISO
      live-iso = nixpkgs.lib.nixosSystem {
        system = "x86_64-linux";
        modules = [
          ./configuration.nix # base do sistema
          "${nixpkgs}/nixos/modules/installer/cd-dvd/installation-cd-minimal.nix"
          ({ ... }: {
            isoImage.squashfsCompression = "gzip -Xcompression-level 1";
          })
        ];
      };

      # VM
      vm-persistente = nixpkgs.lib.nixosSystem {
        system = "x86_64-linux";
        modules = [
          ./configuration.nix # msm base do sistema
          ./vm.nix
          ./hardware-configuration.nix # arq gerado pelo nixos-generate-config na vm
          ({ ... }: {
            # bootloader
            boot.loader.grub.enable = true;
            boot.loader.grub.device = "/dev/vda"; # disco do virt-manager
          })
        ];
      };

    };
  };
}