{ buildGoModule
, lib
}:

buildGoModule {
  name = "go-annotation";

  src = ./.;

  vendorHash = "sha256-S+t0RIzNVQ2kloUo5X7tK1JG1u2T3RuWRINRwHyBu/g=";
}
