class Ccw < Formula
  desc "Claude Code Workspace manager"
  homepage "https://github.com/ccw/ccw"
  version "0.1.15"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.15/ccw-darwin-arm64"
      sha256 "ff7ff31df0808eae04fcddec60b8d7e888c7bfcd91302edce70cd99e05d0358f"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.15/ccw-darwin-amd64"
      sha256 "987c07121999542c924951d2a0b74efb6864b3b21cfc3a31d5e13adf78e7d27b"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.15/ccw-linux-arm64"
      sha256 "927b7963d5a33981d2de8517a10bf74282c0cfff309aca485730a4c76b769f81"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.15/ccw-linux-amd64"
      sha256 "5efdcf335ac4a6367f9854f26d1b5212d7026c90925748ce64f6be25a42a4ff2"
    end
  end

  license "MIT"
  head "https://github.com/ccw/ccw.git", branch: "main"

  def install
    bin.install Dir["ccw*"].first => "ccw"
  end
end
