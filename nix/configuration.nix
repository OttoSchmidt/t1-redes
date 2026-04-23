{ pkgs, ... }:
{
  networking.hostName = "pacman";
  networking.networkmanager.enable = true;
  
  time.timeZone = "America/Sao_Paulo";

  # Configuração de Teclado
  i18n.defaultLocale = "pt_BR.UTF-8";
  console = {
    keyMap = "br-abnt2"; # Para o terminal
  };

  services.xserver = {
    layout = "br";
    xkbVariant = "nodeadkeys"; # Sem teclas mortas
  };

  users.users.pacman = {
    isNormalUser = true;
    description = "Pacman Server";
    extraGroups = [ "wheel" "networkmanager" ];
    initialPassword = "pacman";
  };

  security.sudo.wheelNeedsPassword = false;

  # Servidores para VM
  services.spice-vdagentd.enable = true;
  services.spice-webdavd.enable = true;
  services.qemuGuest.enable = true;

  services.openssh = {
    enable = true;
    settings = {
      PasswordAuthentication = true;
      PermitRootLogin = "no";
    };
  };

  environment.systemPackages = with pkgs; [
    git
    go
    gcc
    gnumake
    libcap
    iproute2
    pciutils
    tcpdump
    vim
    sshfs
  ];

  system.stateVersion = "25.11";
}
