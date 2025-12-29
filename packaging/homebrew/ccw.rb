class Ccw < Formula
  desc "Claude Code Workspace manager"
  homepage "https://github.com/ccw/ccw"
  version "0.1.14"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.14/ccw-darwin-arm64"
      sha256 "2a44321c2c052bf59f5b4258043b266ccafdfea78906c08b581208bb88a0034e"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.14/ccw-darwin-amd64"
      sha256 "f0d016d5636ce4f8c196f4d730672deae801760be678b8fe3ee2525444ecbc90"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.14/ccw-linux-arm64"
      sha256 "9d6035720d0e30b1d26415e843d222ad7502d7d56f87135ba47a5c0c8db6fa91"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.14/ccw-linux-amd64"
      sha256 "915aee87ab20755539ff43159c0efcc083b607ae6165c23b2d96dbfb3fd17380"
    end
  end

  license "MIT"
  head "https://github.com/ccw/ccw.git", branch: "main"

  def install
    bin.install Dir["ccw*"].first => "ccw"
  end
end
