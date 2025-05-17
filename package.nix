{ buildGoModule
, lib
}:

buildGoModule {
  name = "go-annotation";

  src = ./.;

  vendorHash = "sha256-ytvg3WXqREGw0QA1rWsC5+cYdHpk7+d8XIrtmHrmRfY=";
}
