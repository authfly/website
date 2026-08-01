#!/usr/bin/env bun
/**
 * Read a dotted path from fastygo.config.mjs.
 *
 * Usage:
 *   bun scripts/config-get.mjs css.input
 *   bun scripts/config-get.mjs retag.allowHiddenFiles   # one path per line
 *   bun scripts/config-get.mjs retag.baseline
 *   bun scripts/config-get.mjs --json                   # whole config
 */
import config from "../fastygo.config.mjs";

const args = process.argv.slice(2);
if (args.length === 0 || args[0] === "--help") {
  console.error("usage: bun scripts/config-get.mjs <path>|retag.allowHiddenFiles|--json");
  process.exit(2);
}

if (args[0] === "--json") {
  console.log(JSON.stringify(config, null, 2));
  process.exit(0);
}

const key = args[0];

if (key === "retag.allowHiddenFiles") {
  const files = new Set();
  for (const rule of config.retag?.allow ?? []) {
    if (rule.tag === "input" && rule.attrs?.type === "hidden") {
      for (const f of rule.files ?? []) files.add(normalize(f));
    }
  }
  for (const f of [...files].sort()) console.log(f);
  process.exit(0);
}

const value = key.split(".").reduce((acc, part) => (acc == null ? undefined : acc[part]), config);
if (value === undefined) {
  console.error(`config-get: missing key ${key}`);
  process.exit(1);
}
if (typeof value === "object") {
  console.log(JSON.stringify(value));
} else {
  console.log(String(value));
}

function normalize(p) {
  return p.replace(/^\.\//, "").replace(/\\/g, "/");
}
