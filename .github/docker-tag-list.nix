{ buildGoModule
, fetchFromGitHub
, lib
}:

buildGoModule rec {
  pname = "docker-tag-list";
  version = "1.0.1";

  # Don't run tests.
  doCheck = false;
  doInstallCheck = false;

  src = fetchFromGitHub {
    owner = "joejulian";
    repo = pname;
    rev = "v${version}";
    hash = "sha256-iT+GIiO3YQWrOHMD1NoSbwtJLEspCYtKFQcTKicRttc=";
  };

  vendorHash = "sha256-YzDIwLdz6ETZi4y1Eqa8/EizLVqxGirGJCLBmjztNg8=";

  meta = with lib; {
    description = "print lists of docker image tags";
	homepage = "https://github.com/joejulian/docker-tag-list";
    license = licenses.asl20;
    mainProgram = "docker-tag-list";
  };
}
