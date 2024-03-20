{
  inputs = {
    nixpkgs.url = "nixpkgs";
    nixpkgs-unstable.url = "nixpkgs/nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
  };

  outputs =
    inputs@{ self
    , nixpkgs
    , nixpkgs-unstable
    , flake-parts
    ,
    }: flake-parts.lib.mkFlake { inherit inputs; } {
      systems = [ "aarch64-darwin" "x86_64-linux" ];

      perSystem = { system, ... }:
        let
          pkgs = import nixpkgs {
            inherit system;
            overlays = [
              (final: prev: {
                chart-releaser = pkgs.callPackage ./.github/chart-releaser.nix { };
                chart-testing = pkgs.callPackage ./.github/chart-testing.nix { };
              })
            ];
          };
        in
        {
          formatter = pkgs.nixpkgs-fmt;

          devShells.default = pkgs.mkShell {
            buildInputs = [
              pkgs.chart-releaser
              pkgs.chart-testing
              pkgs.dyff
              pkgs.git
              pkgs.go
              pkgs.go-task
              pkgs.helm-docs
              pkgs.kind
              pkgs.kubectl
              pkgs.kubernetes-helm
              pkgs.kustomize
            ];
          };
        };
    };
}
