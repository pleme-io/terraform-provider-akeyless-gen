{
  description = "terraform-provider-akeyless-gen — Generated Terraform provider for Akeyless";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-25.11";
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

      registryOwner = "pleme-io";
      registryName = "akeyless-gen";
      registryDir = "registry.terraform.io/${registryOwner}/${registryName}/${version}/darwin_arm64";

      mkApp = name: script: {
        type = "app";
        program = "${pkgs.writeShellScriptBin name script}/bin/${name}";
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

      apps.${system} = {
        default = {
          type = "app";
          program = "${package}/bin/${pname}";
        };
        build = mkApp "build" ''
          set -euo pipefail
          go build -ldflags "-s -w -X main.version=${version}" -o ${pname}
          echo "built: ./${pname}"
        '';
        install = mkApp "install" ''
          set -euo pipefail
          PLUGIN_DIR="$HOME/.terraform.d/plugins/${registryDir}"
          mkdir -p "$PLUGIN_DIR"
          go build -ldflags "-s -w -X main.version=${version}" -o "$PLUGIN_DIR/${pname}"
          echo "installed to $PLUGIN_DIR"
        '';
        generate = mkApp "generate" ''
          set -euo pipefail
          SPEC="''${SPEC_PATH:-$HOME/code/github/akeylesslabs/akeyless-go/api/openapi.yaml}"
          RESOURCES="''${RESOURCES_PATH:-$HOME/code/github/pleme-io/akeyless-terraform-resources}"
          terraform-forge generate \
            --spec "$SPEC" \
            --resources "$RESOURCES/resources" \
            --output ./internal \
            --provider "$RESOURCES/provider.toml"
        '';
        test = mkApp "test" ''
          set -euo pipefail
          go test ./...
        '';
        check-all = mkApp "check-all" ''
          set -euo pipefail
          echo "=> go vet"
          go vet ./...
          echo "=> go test"
          go test ./...
          echo "=> go build"
          go build -o /dev/null
          echo "done: all checks passed"
        '';
        validate = mkApp "validate" ''
          set -euo pipefail
          SPEC="''${SPEC_PATH:-$HOME/code/github/akeylesslabs/akeyless-go/api/openapi.yaml}"
          RESOURCES="''${RESOURCES_PATH:-$HOME/code/github/pleme-io/akeyless-terraform-resources}"
          terraform-forge validate \
            --spec "$SPEC" \
            --resources "$RESOURCES/resources"
        '';
        drift = mkApp "drift" ''
          set -euo pipefail
          SPEC="''${SPEC_PATH:-$HOME/code/github/akeylesslabs/akeyless-go/api/openapi.yaml}"
          RESOURCES="''${RESOURCES_PATH:-$HOME/code/github/pleme-io/akeyless-terraform-resources}"
          terraform-forge drift \
            --spec "$SPEC" \
            --resources "$RESOURCES/resources"
        '';
        pipeline = mkApp "pipeline" ''
          set -euo pipefail
          echo "=> Step 1: Generate Go code from TOML specs"
          nix run .#generate
          echo ""
          echo "=> Step 2: Build provider"
          go build -ldflags "-s -w -X main.version=${version}" -o ${pname}
          echo ""
          echo "=> Step 3: Run tests"
          go test ./... || true
          echo ""
          echo "done: pipeline complete — ./${pname} ready"
        '';
      };

      formatter.${system} = pkgs.nixfmt-tree;
    };
}
