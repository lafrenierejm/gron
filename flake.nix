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
          version = inputs.self.shortRev or "development";
          gron = pkgs.buildGoApplication {
            inherit version;
            pname = "gron";
            src = ./.;
            ldflags = [ "-X github.com/lafrenierejm/gron/cmd.Version=${version}" ];
            modules = ./gomod2nix.toml;
            meta = with pkgs.lib; {
              description =
                "Transform JSON or YAML into discrete assignments to make it easier to `grep` for what you want and see the absolute 'path' to it";
              homepage = "https://github.com/lafrenierejm/gron";
              license = licenses.mit;
              maintainers = with maintainers; [ lafrenierejm ];
            };
          };
        in {
          # Per-system attributes can be defined here. The self' and inputs'
          # module parameters provide easy access to attributes of the same
          # system.
          packages = {
            inherit gron;
            default = gron;
            gronWithFallback = pkgs.writeShellApplication {
              name = "gron-with-fallback";
              runtimeInputs = [ gron ];
              text = builtins.readFile ./gron-with-fallback.sh;
            };
          };

          apps.default = gron;

          # Auto formatters. This also adds a flake check to ensure that the
          # source tree was auto formatted.
          treefmt.config = {
            projectRootFile = ".git/config";
            package = pkgs.treefmt;
            flakeCheck = false; # use pre-commit's check instead
            programs = {
              gofumpt.enable = true;
              prettier.enable = true;
            };
            settings.formatter = {
              prettier = {
                excludes = [
                  "README.md"
                  "internal/gron/testdata/large-line.json"
                  "internal/gron/testdata/long-stream.json"
                  "internal/gron/testdata/scalar-stream.json"
                  "internal/gron/testdata/stream.json"
                ];
              };
            };
          };

          pre-commit = {
            check.enable = true;
            settings.hooks = {
              editorconfig-checker.enable = true;
              markdownlint.enable = true;
              treefmt.enable = true;
              typos = {
                enable = true;
                excludes = [
                  "ADVANCED.mkd"
                  "internal/gron/testdata/.*"
                  "internal/gron/ungron_test.go"
                ];
              };
            };
          };

          devShells.default = pkgs.mkShell {
            # Inherit all of the pre-commit hooks.
            inputsFrom = [ config.pre-commit.devShell ];
            packages = with pkgs; [
              (mkGoEnv { pwd = ./.; })
              cobra-cli
              go-tools
              godef
              gofumpt
              gomod2nix
              gopls
              gotools
              nixfmt
              nodePackages.prettier
              typos
            ];
          };
        };
    };
}
