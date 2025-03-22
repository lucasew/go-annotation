{ buildGoModule
, lib
}:

buildGoModule {
  name = "go-annotation";

  src = ./.;

  vendorHash = "sha256-3F3T/XLBpGkulsy7a9Rk3qAmrOS5fHKR3VL4jnIBhZs=";
}
