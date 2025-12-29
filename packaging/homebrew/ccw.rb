class Ccw < Formula
  desc "Claude Code Workspace manager"
  homepage "https://github.com/ccw/ccw"
  version "0.1.8"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.8/ccw-darwin-arm64"
      sha256 "3de0d49d462f90d2b719b779cdb53ef8cabeba2a9bb37a3ab5560f04076cc4df"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.8/ccw-darwin-amd64"
      sha256 "9a4fbf77405d14968153503b517a4b9d9ec8361cba49f6949b9baa4f766e006f"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.8/ccw-linux-arm64"
      sha256 "e86be6db7d07071c07bed226c033f5b03451c18b4dcba2985cc38a1a19121076"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.8/ccw-linux-amd64"
      sha256 "80b6defce7ce943040e9bb6bcc26087f1aee0fd87fe56a1ed51b89c430913764"
    end
  end

  license "MIT"
  head "https://github.com/ccw/ccw.git", branch: "main"

  def install
    bin.install Dir["ccw*"].first => "ccw"
  end
end
