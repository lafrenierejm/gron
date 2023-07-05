{
  description = "Nix package, app, and devShell for gron";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/release-23.05";
    flake-parts.url = "github:hercules-ci/flake-parts";
    flake-root.url = "github:srid/flake-root";
    gomod2nix.url = "github:nix-community/gomod2nix";
    pre-commit-hooks-nix = {
      url = "github:cachix/pre-commit-hooks.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = inputs:
    inputs.flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [
        inputs.flake-root.flakeModule
        inputs.pre-commit-hooks-nix.flakeModule
        inputs.treefmt-nix.flakeModule
      ];
      systems = [ "x86_64-linux" "aarch64-darwin" ];
      perSystem = { config, self', inputs', system, ... }:
        let
          pkgs = import inputs.nixpkgs {
            inherit system;
            overlays = [ inputs.gomod2nix.overlays.default (final: prev: { }) ];
            config = { };
          };
        in {
          # Per-system attributes can be defined here. The self' and inputs'
          # module parameters provide easy access to attributes of the same
          # system.
          packages = rec {
            gron = pkgs.buildGoApplication {
              pname = "gron";
              version = self'.shortRev or "dirty";
              # In 'nix develop', we don't need a copy of the source tree
              # in the Nix store.
              src = ./.;
              modules = ./gomod2nix.toml;
              meta = with pkgs.lib; {
                description =
                  "Transform JSON into discrete assignments to make it easier to `grep` for what you want and see the absolute 'path' to it";
                homepage = "https://github.com/tomnomnom/gron";
                license = licenses.mit;
                maintainers = with maintainers; [ lafrenierejm ];
              };
            };
            default = gron;
          };

          # Auto formatters. This also adds a flake check to ensure that the
          # source tree was auto formatted.
          treefmt.config = {
            projectRootFile = ".git/config";
            package = pkgs.treefmt;
            flakeCheck = false; # use pre-commit's check instead
            programs.gofmt.enable = true;
          };

          pre-commit = {
            check.enable = true;
            settings.hooks = {
              markdownlint.enable = true;
              treefmt.enable = true;
              typos = {
                enable = true;
                excludes = [ "ADVANCED.mkd" "testdata/.*" "ungron_test.go" ];
              };
            };
          };

          devShells.default = pkgs.mkShell {
            # Inherit all of the pre-commit hooks.
            inputsFrom = [ config.pre-commit.devShell ];
            buildInputs = with pkgs; [
              go
              godef
              gopls
              gotools
              go-tools
              gomod2nix
              nixfmt
              typos
            ];
          };
        };
    };
}
