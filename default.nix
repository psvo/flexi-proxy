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
pkgs.buildGoModule {
  pname = "flexi-proxy";
  version = "0.1";
  src = ./.;
  vendorHash = "sha256-KI/+s3qVMtqfpb2MpQQsqkng6nNyeXH+H9Km6C9SIbU=";
}
