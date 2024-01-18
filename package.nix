{ buildGoModule
, lib
}:

buildGoModule {
  name = "go-annotation";

  src = ./.;

  vendorHash = "sha256-J+VGeNYlFi1qZQK1QiFTcVZiM1YZYEH959hQVTSA9X0=";
}
