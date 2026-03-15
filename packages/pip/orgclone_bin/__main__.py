"""
Entry point: downloads the correct orgclone binary on first run, then executes it.
"""

import os
import sys
import stat
import platform
import urllib.request
from pathlib import Path

VERSION = "0.1.0"
REPO = "cj-ways/orgclone"
BIN_DIR = Path(__file__).parent / "bin"


def get_binary_name() -> str:
    system = platform.system().lower()
    machine = platform.machine().lower()

    os_map = {"windows": "windows", "darwin": "darwin", "linux": "linux"}
    arch_map = {"x86_64": "amd64", "amd64": "amd64", "aarch64": "arm64", "arm64": "arm64"}

    s = os_map.get(system)
    a = arch_map.get(machine)
    if not s or not a:
        raise RuntimeError(f"Unsupported platform: {system}/{machine}")

    ext = ".exe" if system == "windows" else ""
    return f"orgclone_{s}_{a}{ext}"


def ensure_binary() -> Path:
    BIN_DIR.mkdir(parents=True, exist_ok=True)
    bin_name = get_binary_name()
    bin_path = BIN_DIR / bin_name

    if bin_path.exists():
        return bin_path

    url = f"https://github.com/{REPO}/releases/download/v{VERSION}/{bin_name}"
    print(f"Downloading orgclone {VERSION}...", file=sys.stderr)

    try:
        urllib.request.urlretrieve(url, bin_path)
    except Exception as e:
        print(f"Failed to download binary: {e}", file=sys.stderr)
        print(f"Download manually from: https://github.com/{REPO}/releases", file=sys.stderr)
        sys.exit(1)

    if platform.system() != "Windows":
        bin_path.chmod(bin_path.stat().st_mode | stat.S_IEXEC | stat.S_IXGRP | stat.S_IXOTH)

    return bin_path


def main():
    bin_path = ensure_binary()
    os.execv(str(bin_path), [str(bin_path)] + sys.argv[1:])


if __name__ == "__main__":
    main()
