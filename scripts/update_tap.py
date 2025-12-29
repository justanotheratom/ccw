#!/usr/bin/env python3
"""
Update the Homebrew tap formula with the latest version and checksums.

Usage: update_tap.py <tag> <dist_dir> <formula_path>
- <tag>: release tag (e.g. v0.1.21)
- <dist_dir>: directory containing ccw release binaries (ccw-darwin-amd64, etc.)
- <formula_path>: path to the tap formula file to overwrite

The script computes sha256 for the four binaries and rewrites the formula
with the correct version, URLs, and shas.
"""

import argparse
import hashlib
import os
import sys
from typing import Dict

OWNER = "justanotheratom"
REPO = "ccw"
EXPECTED_FILES = {
    "ccw-darwin-arm64": ("darwin", "arm64"),
    "ccw-darwin-amd64": ("darwin", "amd64"),
    "ccw-linux-arm64": ("linux", "arm64"),
    "ccw-linux-amd64": ("linux", "amd64"),
}


def sha256_file(path: str) -> str:
    h = hashlib.sha256()
    with open(path, "rb") as f:
        for chunk in iter(lambda: f.read(8192), b""):
            h.update(chunk)
    return h.hexdigest()


def compute_shas(dist_dir: str) -> Dict[str, str]:
    shas: Dict[str, str] = {}
    for name in EXPECTED_FILES:
        path = os.path.join(dist_dir, name)
        if not os.path.exists(path):
            raise FileNotFoundError(f"missing release asset: {path}")
        shas[name] = sha256_file(path)
    return shas


def render_formula(version: str, shas: Dict[str, str]) -> str:
    url = f"https://github.com/{OWNER}/{REPO}/releases/download/v{version}/"
    darwin_arm = shas["ccw-darwin-arm64"]
    darwin_amd = shas["ccw-darwin-amd64"]
    linux_arm = shas["ccw-linux-arm64"]
    linux_amd = shas["ccw-linux-amd64"]

    return f"""class Ccw < Formula
  desc "Claude Code Workspace manager"
  homepage "https://github.com/{OWNER}/{REPO}"
  version "{version}"

  on_macos do
    if Hardware::CPU.arm?
      url "{url}ccw-darwin-arm64"
      sha256 "{darwin_arm}"
    else
      url "{url}ccw-darwin-amd64"
      sha256 "{darwin_amd}"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "{url}ccw-linux-arm64"
      sha256 "{linux_arm}"
    else
      url "{url}ccw-linux-amd64"
      sha256 "{linux_amd}"
    end
  end

  license "MIT"
  head "https://github.com/{OWNER}/{REPO}.git", branch: "main"

  def install
    bin.install Dir["ccw*"].first => "ccw"
  end
end
"""


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("tag", help="release tag, e.g. v0.1.21")
    parser.add_argument("dist_dir", help="directory containing release assets")
    parser.add_argument("formula_path", help="path to tap formula")
    args = parser.parse_args()

    tag = args.tag
    if not tag.startswith("v"):
        raise SystemExit("tag must start with 'v' (e.g. v0.1.21)")
    version = tag.lstrip("v")

    shas = compute_shas(args.dist_dir)
    formula = render_formula(version, shas)

    os.makedirs(os.path.dirname(args.formula_path), exist_ok=True)
    with open(args.formula_path, "w", encoding="utf-8") as f:
        f.write(formula)

    return 0


if __name__ == "__main__":
    sys.exit(main())
