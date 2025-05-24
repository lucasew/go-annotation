{ buildGoModule
, lib
}:

buildGoModule {
  name = "go-annotation";

  src = ./.;

  vendorHash = "sha256-oK7qRVEyBkx25rnLwvkjd4yYr1SKehvuS7ooogiKkEU=";
}
