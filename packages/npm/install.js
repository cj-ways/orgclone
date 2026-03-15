#!/usr/bin/env node
/**
 * Postinstall script: downloads the orgclone binary from GitHub Releases.
 * Caches by version in ~/.orgclone/ so reinstalls are instant.
 */

const https = require("https");
const fs = require("fs");
const path = require("path");
const os = require("os");

const REPO = "cj-ways/orgclone";
const VERSION = require("./package.json").binaryVersion;
const IS_WIN = process.platform === "win32";
const BIN_NAME = IS_WIN ? "orgclone.exe" : "orgclone";

// Package bin dir (where npm looks for the executable)
const PKG_BIN = path.join(__dirname, "bin", BIN_NAME);

// Persistent cache dir — survives npm reinstalls
const CACHE_DIR = path.join(os.homedir(), ".orgclone", "bin");
const CACHED_BIN = path.join(CACHE_DIR, `orgclone-${VERSION}${IS_WIN ? ".exe" : ""}`);

function getPlatformTarget() {
  const p = { win32: "windows", darwin: "darwin", linux: "linux" }[process.platform];
  const a = { x64: "amd64", arm64: "arm64" }[process.arch];
  if (!p || !a) throw new Error(`Unsupported platform: ${process.platform}/${process.arch}`);
  return `orgclone_${p}_${a}${IS_WIN ? ".exe" : ""}`;
}

function download(url, dest) {
  return new Promise((resolve, reject) => {
    const follow = (u) => {
      https.get(u, { headers: { "User-Agent": "orgclone-installer" } }, (res) => {
        if (res.statusCode === 301 || res.statusCode === 302) return follow(res.headers.location);
        if (res.statusCode !== 200) return reject(new Error(`HTTP ${res.statusCode}`));
        const file = fs.createWriteStream(dest);
        res.pipe(file);
        file.on("finish", () => file.close(resolve));
        file.on("error", reject);
      }).on("error", reject);
    };
    follow(url);
  });
}

async function main() {
  fs.mkdirSync(CACHE_DIR, { recursive: true });
  fs.mkdirSync(path.join(__dirname, "bin"), { recursive: true });

  // If this version is already cached, just copy it — no download needed
  if (fs.existsSync(CACHED_BIN)) {
    fs.copyFileSync(CACHED_BIN, PKG_BIN);
    if (!IS_WIN) fs.chmodSync(PKG_BIN, 0o755);
    console.log(`orgclone ${VERSION} ready.`);
    return;
  }

  // Download and cache
  const url = `https://github.com/${REPO}/releases/download/v${VERSION}/${getPlatformTarget()}`;
  console.log(`Downloading orgclone ${VERSION}...`);

  try {
    await download(url, CACHED_BIN);
    if (!IS_WIN) fs.chmodSync(CACHED_BIN, 0o755);
    fs.copyFileSync(CACHED_BIN, PKG_BIN);
    if (!IS_WIN) fs.chmodSync(PKG_BIN, 0o755);
    console.log("orgclone installed successfully.");
  } catch (err) {
    console.error(`Failed to download binary: ${err.message}`);
    console.error(`Download manually from: https://github.com/${REPO}/releases`);
    process.exit(1);
  }
}

main();
