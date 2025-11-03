{
  pkgs,
  name,
  self,
  buildGo,
  version,
  ...
}:
with pkgs;
let
  concat = y: x: y + "/" + x;
  namespacedName = concat name;
  globalVersion = version;
  globalVendorHash = "sha256-yCoCskJH7zC8qrPSPTustmfUInJGJQPW406UlTvYKik=";
  # globalVendorHash = lib.fakeHash;

  customGoBuild =
    {
      name,
      version ? globalVersion,
      subPackages ? [ ],
    }:
    let
      fullName = namespacedName name;
      binary = buildGo {
        pname = fullName;
        version = version;
        src = lib.cleanSource self;
        vendorHash = globalVendorHash;
        inherit subPackages;
        doCheck = false;

        ldflags = [
          "-s -w -X main.Version=${version}"
        ];

        env = {
          CGO_ENABLED = false;
        };

        tags = [
          "netgo"
          "osusergo"
        ];
      };

      docker = dockerTools.buildLayeredImage {
        name = fullName;
        tag = version;
        contents = [ binary ];
        config = {
          Entrypoint = [ "${binary}/bin/server" ];
          Env = [
            "PATH=${binary}/bin"
          ];
        };
      };
    in
    {
      binary = binary;
      docker = docker;
    };
  api = customGoBuild {
    name = "api";
    subPackages = [
      "apps/api/cmd/server"
      "apps/api/cmd/healthcheck"
    ];
  };

  notify = customGoBuild {
    name = "notify";
    subPackages = [
      "apps/notify/cmd/server"
    ];
  };

in
{
  default = api.docker;

  api-bin = api.binary;
  notify-bin = notify.binary;

  api-docker = api.docker;
  notify-docker = notify.docker;
}
