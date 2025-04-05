{ buildGoModule
, lib
}:

buildGoModule {
  name = "go-annotation";

  src = ./.;

  vendorHash = "sha256-myGjgPSWiZtvi2Qaxiytp5zBZS9YoGFI+q+v/VTqS28=";
}
