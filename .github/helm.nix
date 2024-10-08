{ pkgs
, stdenv
, fetchzip
}:
let
  pname = "helm";
  version = "3.10.3";
  src = {
    aarch64-darwin = fetchzip {
      url = "https://get.helm.sh/helm-v${version}-darwin-arm64.tar.gz";
      hash = "sha256-3W/piPZvkyrGOLCgghn7j9CgNxAVvWn1kwFb8Von9Ko=";
    };
    x86_64-linux = fetchzip {
      url = "https://get.helm.sh/helm-v${version}-linux-amd64.tar.gz";
      hash = "sha256-XAtiT7vaSBrfrj03gbcQUmUMQSZ9+5nymxfVSOnQ+sM=";
    };
  }.${stdenv.system} or (throw "${pname}-${version}: ${stdenv.system} is unsupported.");
in
stdenv.mkDerivation {
  inherit pname version src;

  installPhase = ''
    mkdir -p "$out/bin"
    cp "$src/helm" "$out/bin/helm-${version}"
    chmod 755 "$out/bin/helm-${version}"
  '';
}
