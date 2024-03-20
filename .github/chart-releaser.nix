# Pulled from https://github.com/NixOS/nixpkgs/blob/fa9f817df522ac294016af3d40ccff82f5fd3a63/pkgs/applications/networking/cluster/helm/chart-testing/default.nix#L62
# and adapted to use https://github.com/helm/chart-releaser
{ buildGoModule
, coreutils
, fetchFromGitHub
, git
, installShellFiles
, lib
, makeWrapper
}:

buildGoModule rec {
  pname = "chart-releaser";
  version = "1.6.0";

  # Don't run tests.
  doCheck = false;
  doInstallCheck = false;

  src = fetchFromGitHub {
    owner = "helm";
    repo = pname;
    rev = "v${version}";
    hash = "sha256-rPNGg4nrDFIa1PAw3efFU/pQub33+QD0vNFu8kiU2/E=";
  };

  vendorHash = "sha256-zBVAER1RJy449GUndvQkG8R84vOuL+IN4exjETVHp9k=";

  postPatch = ''
    substituteInPlace pkg/config/config.go \
      --replace "\"/etc/cr\"," "\"$out/etc/cr\","
  '';

  # https://github.com/helm/chart-releaser/blob/fa01315c4668d4fca627a5afc67409e31b27305c/.goreleaser.yml#L37
  ldflags = [
    "-w"
    "-s"
    "-X github.com/helm/chart-releaser/cr/cmd.Version=${version}"
    "-X github.com/helm/chart-releaser/cr/cmd.GitCommit=${src.rev}"
    "-X github.com/helm/chart-releaser/cr/cmd.BuildDate=19700101-00:00:00"
  ];

  nativeBuildInputs = [ installShellFiles makeWrapper ];

  postInstall = ''
     installShellCompletion --cmd cr \
       --bash <($out/bin/cr completion bash) \
       --zsh <($out/bin/cr completion zsh) \
       --fish <($out/bin/cr completion fish) \

    wrapProgram $out/bin/cr --prefix PATH : ${lib.makeBinPath [
      coreutils
      git
    ]}
  '';

  meta = with lib; {
    description = "Hosting Helm Charts via GitHub Pages and Releases";
    homepage = "https://github.com/helm/chart-releaser";
    license = licenses.asl20;
    mainProgram = "cr";
  };
}
