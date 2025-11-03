{
  pkgs,
}:
with pkgs;
mkShell {
  name = "go shell";

  buildInputs = with pkgs; [
    go
    golangci-lint
    golangci-lint-langserver

    go-swag
    openapi-generator-cli

    # LSPs
    nixd
    vscode-langservers-extracted # json lsp
    delve
    gopls
    yaml-language-server
    docker-compose-language-service
  ];

  shellHook = ''
    git config core.hooksPath .hooks
  '';
}
