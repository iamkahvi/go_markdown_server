{
  description = "Dev shell for go_markdown_server";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }:
    let
      lib = nixpkgs.lib;
      systems = [
        "aarch64-darwin"
        "x86_64-darwin"
        "aarch64-linux"
        "x86_64-linux"
      ];
    in {
      devShells = lib.genAttrs systems (system:
        let
          pkgs = import nixpkgs {
            inherit system;
          };
        in {
          default = pkgs.mkShell {
            packages = with pkgs; [
              go
              delve
              gopls
              golangci-lint
            ];

            shellHook = ''
              export GOPATH="$PWD/.gopath"
              export GOBIN="$GOPATH/bin"
              export PATH="$GOBIN:$PATH"
							export DEV=1
              mkdir -p "$GOBIN"
							alias s="go run ./cmd/server"
            '';
          };
        });
    };
}
