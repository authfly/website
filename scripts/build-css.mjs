#!/usr/bin/env bun
/** Build Tailwind CSS using paths from fastygo.config.mjs. */
import { $ } from "bun";
import config from "../fastygo.config.mjs";

const input = config.css.input;
const output = config.css.output;
await $`bunx @tailwindcss/cli -i ${input} -o ${output} --minify`;
