class Ccw < Formula
  desc "Claude Code Workspace manager"
  homepage "https://github.com/ccw/ccw"
  version "0.1.1"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.1/ccw-darwin-arm64"
      sha256 "320729b3e16f3b2260ff821c0116aa75aaf1072e1cf45613f3bd362d1900bbd5"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.1/ccw-darwin-amd64"
      sha256 "bc9326805c5e16ff7c750b1294e098c28d0c9af115a2737c1772d6849b07816e"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.1/ccw-linux-arm64"
      sha256 "76e3375a113c34bf32da2df642bd40e713bafac242755e840dfd8961665a5e38"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.1/ccw-linux-amd64"
      sha256 "d5d75ee9b8676b0139f6d61c02da5e84c22c6e214e6995a217eac1745c059d50"
    end
  end

  license "MIT"
  head "https://github.com/ccw/ccw.git", branch: "main"

  def install
    bin.install Dir["ccw*"].first => "ccw"
  end
end
