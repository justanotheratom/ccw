class Ccw < Formula
  desc "Claude Code Workspace manager"
  homepage "https://github.com/ccw/ccw"
  version "0.1.8"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.8/ccw-darwin-arm64"
      sha256 "b77e630b6ad31bdb15887be10ac2b4b735d0db23d304910d4ed7eb15d74f0b5d"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.8/ccw-darwin-amd64"
      sha256 "9a4fbf77405d14968153503b517a4b9d9ec8361cba49f6949b9baa4f766e006f"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.8/ccw-linux-arm64"
      sha256 "493590435ba1650652196e36c4110efba316512809230effe753c4be6060533e"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.8/ccw-linux-amd64"
      sha256 "c743595c1db1929b2391eb5acb20f52a8e95afa66aa0c1604e139373bd2a06f4"
    end
  end

  license "MIT"
  head "https://github.com/ccw/ccw.git", branch: "main"

  def install
    bin.install Dir["ccw*"].first => "ccw"
  end
end
