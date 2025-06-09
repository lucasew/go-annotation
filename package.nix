{ buildGoModule
, lib
, self ? {}
}:

buildGoModule {
  pname = "go-annotation";
  version = "${builtins.readFile ./version.txt}-${self.shortRev or self.dirtyShortRev or "rev"}";

  src = ./.;

  vendorHash = "sha256-mIu8b7v6X08+cyzc3zf80zQINQcsRBDesDG9bpaGw3U=";
}
