{
  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
  };

  outputs =
    inputs@{ self
    , nixpkgs
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
                helm-3-10-3 = pkgs.callPackage ./.github/helm.nix { };
                setup-envtest = pkgs.callPackage ./.github/setup-envtest.nix { };
                kubernetes-helm = prev.wrapHelm prev.kubernetes-helm {
                  plugins = [ prev.kubernetes-helmPlugins.helm-unittest ];
                };
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
              pkgs.gnutar
              pkgs.go
              pkgs.go-task
              pkgs.gofumpt
              pkgs.helm-3-10-3
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
