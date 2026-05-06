{ pkgs, ... }:
{
  fileSystems."/home/pacman/pacman" = {
    device = "pacman";
    fsType = "virtiofs";
  };
}