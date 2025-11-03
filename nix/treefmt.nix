{ pkgs, ... }:
{
  projectRootFile = "flake.nix";
  programs.gofmt.enable = true;
  programs.prettier.enable = true;
  programs.nixfmt.enable = true;
  programs.sqlfluff.enable = true;
  programs.sqlfluff.dialect = "postgres";
  programs.yamlfmt.enable = true;
  settings.global.excludes = [
    "pnpm-lock.yaml"
    "apps/api/internal/database/**"
  ];
}
