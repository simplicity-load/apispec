{
  description = "ApiSpec Flake";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-25.05";
    nixpkgs-unstable.url = "nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    treefmt-nix.url = "github:numtide/treefmt-nix";
  };

  outputs =
    {
      self,
      nixpkgs,
      nixpkgs-unstable,
      flake-utils,
      treefmt-nix,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        name = "apispec";
        version = "0.0.1";
        pkgs = nixpkgs.legacyPackages.${system};
        unstable = nixpkgs-unstable.legacyPackages.${system};
        treefmtEval = treefmt-nix.lib.evalModule pkgs ./nix/treefmt.nix;
        buildGo = pkgs.buildGo124Module;
      in
      {
        devShells.default = import ./nix/shell.nix {
          inherit pkgs;
        };
        packages = import ./nix/packages.nix {
          inherit
            pkgs
            name
            self
            buildGo
            version
            ;
        };
        apps = import ./nix/apps.nix {
          inherit pkgs;
        };
        formatter = treefmtEval.config.build.wrapper;
        checks = {
          formatting = treefmtEval.config.build.check self;
        };
      }
    );
}
