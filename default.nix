{ nixpkgs ?
  # Same as leksah
        import ((import <nixpkgs> {}).pkgs.fetchFromGitHub {
          owner = "NixOS"; repo = "nixpkgs";
          rev = "f8e8ecde51b49132d7f8d5adb971c0e37eddcdc2";
          sha256 = "14b7945442q5hlhvhnm15y3cds2lmm6kn52srv2bbr3yla6b2pv9";
        }) {}
}:

nixpkgs.buildGoPackage rec {
  name = "awsRedisServer";
  version = "0.1.0";
  # buildInputs = [ net osext ];
  # src = ./.;
  goPackagePath = "github.com/dfordivam/awsRedisServer";
  goDeps = ./deps.nix;
}