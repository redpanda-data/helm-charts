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
              })
            ];
          };
        in
        {
          formatter = pkgs.nixpkgs-fmt;

          packages =
            let
              buildCmdGoModule = { package, srcFilters }: pkgs.buildGoModule {
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
                src = pkgs.lib.sources.sourceByRegex ./. ([
                  # Including by regex requires that all folders within the
                  # hierarchy are matched. IE To include
                  # `charts/redpanda/foo.go` There must be a match for
                  # `charts`, `charts/redpanda`, and `charts/redpanda/foo.go`
                  "^go.mod$"
                  "^go.sum$"
                ] ++ srcFilters);

                # Effectively a nix lock file for go.mod and go.sum to know if
                # deps need to be re-downloaded.
                # To update:
                # 1. set to and empty string
                # 2. Run `nix develop`
                # 3. Copy the output value into vendorHash
                # TODO: Figure out a better way to update this.
                vendorHash = "sha256-tIFpf0UcsI7vmQZGfTKOzGH/6gM+FYiTt2foojLgOjY=";
              };
            in
            {
              gotohelm = buildCmdGoModule {
                package = "gotohelm";
                srcFilters = [
                  "^cmd$"
                  "^cmd/gotohelm(/.*)?$"
                  "^pkg(/.*)?$"
                ];
              };
              genpartial = buildCmdGoModule {
                package = "genpartial";
                srcFilters = [
                  "^cmd$"
                  "^cmd/genpartial(/.*)?$"
                  "^pkg(/.*)?$"
                ];
              };
              genschema = buildCmdGoModule {
                package = "genschema";
                srcFilters = [
                  "^cmd(/.*)?$"
                  "^pkg(/.*)?$"
                  # genschema currently needs to import the redpanda package as
                  # it uses reflection.
                  "^charts$"
                  "^charts/redpanda$"
                  "^charts/.*\.go$"
                  "^charts/.*\.schema.json$"
                ];
              };
            };

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
              pkgs.yq # jq but for YAML
            ];
          };
        };
    };
}
