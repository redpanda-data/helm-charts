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
              })
            ];
          };
        in
        {
          formatter = pkgs.nixpkgs-fmt;

          packages =
            let
              buildCmdGoModule = package: pkgs.buildGoModule {
                pname = package;
                version = "0.0.0";

                # Don't run tests.
                doCheck = false;
                doInstallCheck = false;

                # Only compile this specific binary.
                subPackages = [ "cmd/${package}" ];

                # All files required for running `go build`
                # We filter out other extraneous files so `nix develop` doesn't
                # rebuild our go programs needlessly.
                src = pkgs.lib.sources.sourceByRegex ./. [
                  # Including by regex requires that all folders within the
                  # hierarchy are matched. IE To include
                  # `charts/redpanda/foo.go` There must be a match for
                  # `charts`, `charts/redpanda`, and `charts/redpanda/foo.go`
                  "^go.mod$"
                  "^go.sum$"
                  "^cmd(/.*)?$"
                  "^pkg(/.*)?$"
                  "^charts$"
                  "^charts/redpanda$"
                  "^charts(/.*\.go)?$"
                ];

                # Effectively a nix lock file for go.mod and go.sum to know if
                # deps need to be re-downloaded.
                # To update:
                # 1. set to and empty string
                # 2. Run `nix develop`
                # 3. Copy the output value into vendorHash
                # TODO: Figure out a better way to update this.
                vendorHash = "sha256-ILhTJL3KBuZNXpCLxbWFGK7+9cUb2imPHjgfTxqnkjM=";
              };
            in
            {
              gotohelm = buildCmdGoModule "gotohelm";
              genpartial = buildCmdGoModule "genpartial";
              genvalues = buildCmdGoModule "genvalues";
            };

          devShells.default = pkgs.mkShell {
            buildInputs = [
              pkgs.actionlint # Github Workflow definition linter https://github.com/rhysd/actionlint
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
