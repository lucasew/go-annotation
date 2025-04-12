{ buildGoModule
, lib
}:

buildGoModule {
  name = "go-annotation";

  src = ./.;

  vendorHash = "sha256-2kM5CfGnLgAY+gx3oh4HjgFQ13NYkm3cECZ6yqBlZVY=";
}
