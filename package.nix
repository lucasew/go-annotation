{ buildGoModule
, lib
}:

buildGoModule {
  name = "go-annotation";

  src = ./.;

  vendorHash = "sha256-Qeu8WQEblZ9LVnVu1dWI20Ssgk2EmFwNMc1SiOOXqrc=";
}
