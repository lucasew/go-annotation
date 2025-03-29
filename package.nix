{ buildGoModule
, lib
}:

buildGoModule {
  name = "go-annotation";

  src = ./.;

  vendorHash = "sha256-jWid76GJN+6Ta7PK0Uy9kTQn20+2lCqM2sb4TakbceQ=";
}
