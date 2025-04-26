{ buildGoModule
, lib
}:

buildGoModule {
  name = "go-annotation";

  src = ./.;

  vendorHash = "sha256-/VwQKNt0kzGjc1lkcU/jcthH3vJuQtp0i5eBoYj5A3U=";
}
