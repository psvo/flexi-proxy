{
  pkgs ? (
    let
      inherit (builtins) fetchTree fromJSON readFile;
      inherit ((fromJSON (readFile ./flake.lock)).nodes) nixpkgs;
    in
      import (fetchTree nixpkgs.locked) {
        overlays = [
          #(import "${fetchTree gomod2nix.locked}/overlay.nix")
        ];
      }
  ),
}:
pkgs.mkShell {
  nativeBuildInputs = with pkgs; [
    go
    go-outline
    gopls
    gopkgs
    go-tools
    delve
  ];
  hardeningDisable = ["all"];
}
