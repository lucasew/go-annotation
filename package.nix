{ buildGoModule
, lib
}:

buildGoModule {
  name = "go-annotation";

  src = ./.;

  vendorHash = "sha256-XP8dCCzp7wZjo/iXogGoygJ4DurNzi0banN1B7gmLqc=";
}
