{ buildGoModule
, lib
}:

buildGoModule {
  name = "go-annotation";

  src = ./.;

  vendorHash = "sha256-ZdhobWOrSdtFJX92QtU/16v1F2JWU7N/aNVhW9jvO20=";
}
