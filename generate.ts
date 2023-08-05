#!/usr/bin/env -S deno run -A
import { doc } from "https://deno.land/x/deno_doc@0.64.0/mod.ts";
import { parse } from "https://deno.land/std@0.166.0/flags/mod.ts";

const flags = parse(Deno.args, {
  string: ["api-version", "repo-version"],
  default: {
    "api-version": "10",
    "repo-version": "0.37.51",
  },
});

function requireFlag(name: string): string | number {
  const value = flags[name];
  if (value == undefined) {
    const flag = name ? `--${name}` : `-${name}`;
    throw new Error("Missing required flag: " + flag);
  }
  return value as string | number;
}

const apiVersion = requireFlag("api-version");
const repoVersion = requireFlag("repo-version");

const discordTypesURL = (() => {
  const versionBit = repoVersion && `@${repoVersion}`;
  return `https://deno.land/x/discord_api_types${versionBit}/v${apiVersion}.ts`;
})();

async function main() {
  const discordTypes = await doc(discordTypesURL);
  for (const node of discordTypes) {
    console.log(`name: ${node.name} kind: ${node.kind}`);
  }
}

await main();
