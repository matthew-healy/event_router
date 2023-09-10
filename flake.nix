{
  description = "playing with event routing in Go";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    pre-commit-hooks = {
      url = "github:cachix/pre-commit-hooks.nix";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.flake-utils.follows = "flake-utils";
    };
  };

  outputs = { self, nixpkgs, flake-utils, pre-commit-hooks }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };

        preCommitHook =
          ({ doFullLint ? false }: pre-commit-hooks.lib.${system}.run {
            src = self;

            hooks = {
              nixpkgs-fmt.enable = true;

              gofmt.enable = true;
              gotest.enable = true;
            };
          });

        eventRouterPkg = pkgs.buildGoModule {
          pname = "event_router";
          version = "v0";

          src = ./.;

          vendorHash = null;
          # vendorSha256 = pkgs.lib.fakeSha256;
        };
      in
      {
        packages = rec {
          event_router = eventRouterPkg;

          default = event_router;
        };

        formatter = pkgs.nixpkgs-fmt;

        devShells.default = pkgs.mkShell {
          shellHook = ''
            ${(preCommitHook {}).shellHook}
          '';

          nativeBuildInputs = with pkgs; [
            go
            gopls
            golangci-lint
          ];
        };

        checks = {
          event_router = eventRouterPkg;
          pre-commit = preCommitHook { doFullLint = true; };
        };
      }
    );
}
