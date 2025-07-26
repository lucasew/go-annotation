{ buildGoModule
, lib
, self ? {}
}:

buildGoModule {
  pname = "go-annotation";
  version = "${builtins.readFile ./version.txt}-${self.shortRev or self.dirtyShortRev or "rev"}";

  src = ./.;

  vendorHash = "sha256-B9s2D2rPGt794zr5fFuNIAZMOH/31Oluy3kNUTWLJo8=";
}
