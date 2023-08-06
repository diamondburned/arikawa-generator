{ pkgs ? import <nixpkgs> {} }:

let
	sources = import ./nix/sources.nix {};
in

pkgs.mkShell {
	DISCORD_API_DOCS = sources.discord-api-docs;
	DISCORD_API_SPEC = sources.discord-api-spec;

	buildInputs = with pkgs; [
		go
		niv
	];
}
