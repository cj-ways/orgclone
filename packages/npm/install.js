#!/usr/bin/env node
/**
 * Postinstall script: downloads the correct orgclone binary from GitHub Releases.
 */

const https = require("https");
const fs = require("fs");
const path = require("path");
const { execSync } = require("child_process");
const zlib = require("zlib");

const REPO = "cj-ways/orgclone";
const VERSION = require("./package.json").version;
const BIN_DIR = path.join(__dirname, "bin");
const BIN_PATH = path.join(BIN_DIR, process.platform === "win32" ? "orgclone.exe" : "orgclone");

function getPlatformTarget() {
  const platform = process.platform;
  const arch = process.arch;

  const platformMap = { win32: "windows", darwin: "darwin", linux: "linux" };
  const archMap = { x64: "amd64", arm64: "arm64" };

  const p = platformMap[platform];
  const a = archMap[arch];

  if (!p || !a) throw new Error(`Unsupported platform: ${platform}/${arch}`);

  const ext = platform === "win32" ? ".exe" : "";
  return `orgclone_${p}_${a}${ext}`;
}

function download(url, dest) {
  return new Promise((resolve, reject) => {
    const follow = (u) => {
      https.get(u, { headers: { "User-Agent": "orgclone-installer" } }, (res) => {
        if (res.statusCode === 301 || res.statusCode === 302) {
          follow(res.headers.location);
          return;
        }
        if (res.statusCode !== 200) {
          reject(new Error(`HTTP ${res.statusCode} for ${u}`));
          return;
        }
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
  if (fs.existsSync(BIN_PATH)) {
    console.log("orgclone binary already installed.");
    return;
  }

  fs.mkdirSync(BIN_DIR, { recursive: true });

  const target = getPlatformTarget();
  const url = `https://github.com/${REPO}/releases/download/v${VERSION}/${target}`;

  console.log(`Downloading orgclone ${VERSION} for ${process.platform}/${process.arch}...`);

  try {
    await download(url, BIN_PATH);
    if (process.platform !== "win32") {
      fs.chmodSync(BIN_PATH, 0o755);
    }
    console.log("orgclone installed successfully.");
  } catch (err) {
    console.error(`Failed to download binary: ${err.message}`);
    console.error(`You can manually download from: https://github.com/${REPO}/releases`);
    process.exit(1);
  }
}

main();
