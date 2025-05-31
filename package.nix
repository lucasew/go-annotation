{ buildGoModule
, lib
}:

buildGoModule {
  name = "go-annotation";

  src = ./.;

  vendorHash = "sha256-+FfmyWepfbH6GBzVXo6J12rHgeMU8m0lVMMqToowSSw=";
}
