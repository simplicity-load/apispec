{
  pkgs,
  ...
}:
let
  goLint = pkgs.writeShellScriptBin "golint" ''
    ${pkgs.golangci-lint}/bin/golangci-lint run --issues-exit-code 0
  '';
in
rec {
  go_lint = {
    name = "lint go";
    type = "app";
    program = "${goLint}/bin/golint";
  };
  default = go_lint;
}
