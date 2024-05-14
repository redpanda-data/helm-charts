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
    , ...
    }: flake-parts.lib.mkFlake { inherit inputs; } {
      systems = [ "aarch64-darwin" "x86_64-linux" ];

      perSystem = { self', system, ... }:
        let
          pkgs = import nixpkgs {
            inherit system;
            overlays = [
              (final: prev: {
                chart-releaser = pkgs.callPackage ./.github/chart-releaser.nix { };
                chart-testing = pkgs.callPackage ./.github/chart-testing.nix { };
                docker-tag-list = pkgs.callPackage ./.github/docker-tag-list.nix { };
                setup-envtest = pkgs.callPackage ./.github/setup-envtest.nix { };
              })
            ];
          };
        in
        {
          formatter = pkgs.nixpkgs-fmt;

          devShells.default = pkgs.mkShell {
            buildInputs = [
              pkgs.actionlint # Github Workflow definition linter https://github.com/rhysd/actionlint
              pkgs.chart-releaser
              pkgs.chart-testing
              pkgs.docker-tag-list # Utility to list out docker tags
              pkgs.dyff
              pkgs.gh # Github CLI
              pkgs.git
              pkgs.go
              pkgs.go-task
              pkgs.helm-docs
              pkgs.jq # CLI JSON swiss army knife
              pkgs.kind
              pkgs.kube-linter
              pkgs.kubectl
              pkgs.kubernetes-helm
              pkgs.kustomize
              pkgs.setup-envtest
              pkgs.yq # jq but for YAML
            ];
          };
        };
    };
}
