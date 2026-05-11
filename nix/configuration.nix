{ pkgs, ... }:
{
  networking.hostName = "pacman";
  networking.networkmanager.enable = true;
  
  time.timeZone = "America/Sao_Paulo";

  # Configuração de Teclado
  i18n.defaultLocale = "pt_BR.UTF-8";
  console = {
    keyMap = "br-abnt2"; # Para o terminal
    font = "ter-v16n";
    packages = with pkgs; [ terminus_font ];
  };

  services.xserver = {
    enable = true;
    displayManager.gdm.enable = true;
    desktopManager.gnome.enable = true;
    layout = "br";
    xkbVariant = "nodeadkeys"; # Sem teclas mortas
  };

  # remover console padrao, pois teremos o ptyxis
  environment.gnome.excludePackages = with pkgs; [
    gnome-console
  ];

  fonts = {
    fontconfig.enable = true;
    enableDefaultPackages = true;
    packages = with pkgs; [
      dejavu_fonts
      liberation_ttf
      noto-fonts
      noto-fonts-cjk-sans
      noto-fonts-color-emoji
    ];
    fontconfig.defaultFonts = {
      monospace = [ "DejaVu Sans Mono" "Noto Sans Mono" ];
      sansSerif = [ "DejaVu Sans" "Noto Sans" ];
      serif = [ "DejaVu Serif" "Noto Serif" ];
      emoji = [ "Noto Color Emoji" ];
    };
  };

  users.users.pacman = {
    isNormalUser = true;
    description = "Pacman Server";
    extraGroups = [ "wheel" "networkmanager" "docker" ];
    initialPassword = "pacman";
  };

  # auto login
  services.displayManager.autoLogin.enable  = true;
  services.displayManager.autoLogin.user = "pacman";

  # sem senha sudo (grupo wheel)
  security.sudo.wheelNeedsPassword = false;

  # servidores para vm (transferencia de arquivos)
  services.spice-vdagentd.enable = true;
  services.spice-webdavd.enable = true;
  services.qemuGuest.enable = true;

  # docker
  virtualisation.docker.enable = true;

  programs.wireshark.enable = true;

  services.openssh = {
    enable = true;
    settings = {
      PasswordAuthentication = true;
      PermitRootLogin = "no";
    };
  };

  nixpkgs.config.allowUnfree = true;
  environment.systemPackages = with pkgs; [
    git
    go
    gcc
    gnumake
    libcap
    iproute2
    pciutils
    ptyxis
    tcpdump
    vim
    vscode
    sshfs
    wireshark
  ];

  system.stateVersion = "25.11";
}
