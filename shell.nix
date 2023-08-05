{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
	buildInputs = with pkgs; [ deno ];

	DENO_NO_UPDATE_CHECK = "1";
}
