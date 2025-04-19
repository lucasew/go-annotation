{ buildGoModule
, lib
}:

buildGoModule {
  name = "go-annotation";

  src = ./.;

  vendorHash = "sha256-BE+e4C/qwSjVRUIWzrPq6XBM6R/l3P3u9v1Uh9EIqM0=";
}
