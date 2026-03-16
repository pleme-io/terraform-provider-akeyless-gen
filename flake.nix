{
  description = "terraform-provider-akeyless-gen — Generated Terraform provider for Akeyless";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    substrate = {
      url = "github:pleme-io/substrate";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      substrate,
      ...
    }:
    let
      system = "aarch64-darwin";
      pkgs = import nixpkgs { inherit system; };

      version = "0.1.0";
      pname = "terraform-provider-akeyless-gen";

      package = pkgs.buildGoModule {
        inherit pname version;
        src = pkgs.lib.cleanSource ./.;
        vendorHash = null;
        ldflags = [
          "-s"
          "-w"
          "-X main.version=${version}"
        ];
        doCheck = true;
        meta = {
          description = "Generated Terraform provider for Akeyless (terraform-plugin-framework)";
          homepage = "https://github.com/pleme-io/terraform-provider-akeyless-gen";
          license = pkgs.lib.licenses.mit;
          mainProgram = pname;
        };
      };
    in
    {
      packages.${system} = {
        ${pname} = package;
        default = package;
      };

      overlays.default = final: prev: {
        ${pname} = self.packages.${final.system}.default;
      };

      devShells.${system}.default = pkgs.mkShellNoCC {
        packages = [
          pkgs.go
          pkgs.gopls
          pkgs.gotools
          pkgs.terraform
        ];
      };

      formatter.${system} = pkgs.nixfmt-tree;
    };
}
