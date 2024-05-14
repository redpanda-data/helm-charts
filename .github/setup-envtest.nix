{ buildGoModule
, fetchFromGitHub
, lib
}:

buildGoModule rec {
  pname = "setup-envtest";
  version = "0.18.2";

  # Don't run tests.
  doCheck = false;
  doInstallCheck = false;

  src = fetchFromGitHub {
    owner = "kubernetes-sigs";
    repo = "controller-runtime";
    rev = "v${version}";
    hash = "sha256-fQgWwndxzBIi3zsNMYvFDXjetnaQF0NNK+qW8j4Wn/M=";
  };

  sourceRoot = "source/tools/setup-envtest";

  vendorHash = "sha256-Xr5b/CRz/DMmoc4bvrEyAZcNufLIZOY5OGQ6yw4/W9k=";

  meta = with lib; {
    description = "A small tool that manages binaries for envtest";
    homepage = "https://github.com/kubernetes-sigs/controller-runtime/tree/main/tools/setup-envtest";
    license = licenses.asl20;
    mainProgram = "setup-envtest";
  };
}
