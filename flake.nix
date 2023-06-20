{
  description = "Flexi Proxy flake";

  inputs.nixpkgs.url = "flake:nixpkgs";
  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }: (
    flake-utils.lib.eachDefaultSystem
    (system: let
      pkgs = import nixpkgs {
        inherit system;
        overlays = [];
      };
    in {
      # Use alejandra for `nix fmt'
      formatter = pkgs.alejandra;
      packages.default = pkgs.callPackage ./. {};
      devShells.default = import ./shell.nix {inherit pkgs;};
    })
  );
}
