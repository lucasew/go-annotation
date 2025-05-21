{ buildGoModule
, lib
}:

buildGoModule {
  name = "go-annotation";

  src = ./.;

  vendorHash = "sha256-FuuaX+dnF6sCII8qhqbrmplK9p464qPh8M/qpjEXSZA=";
}
