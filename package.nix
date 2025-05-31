{ buildGoModule
, lib
, self ? {}
}:

buildGoModule {
  pname = "go-annotation";
  version = "${builtins.readFile ./version.txt}-${self.shortRev or self.dirtyShortRev or "rev"}";

  src = ./.;

  vendorHash = "sha256-+FfmyWepfbH6GBzVXo6J12rHgeMU8m0lVMMqToowSSw=";
}
