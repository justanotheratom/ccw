class Ccw < Formula
  desc "Claude Code Workspace manager"
  homepage "https://github.com/ccw/ccw"
  version "0.1.20"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.20/ccw-darwin-arm64"
      sha256 "8f5d0c17a845260c9ca1e60a4c00d409d6eb067a1121d702177e6746deebfae5"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.20/ccw-darwin-amd64"
      sha256 "584c8ade41052b57769f185771444714e3005b19141c901dcef2b561b75adabc"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.20/ccw-linux-arm64"
      sha256 "d4939065120e44bfc4b85643ec7c815a892ba04d733040dbbfe5dca27bc9c930"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.20/ccw-linux-amd64"
      sha256 "8447bfe348c64bd3e26744788bd9582d55ace495b13b8b78dd927467bbff5f38"
    end
  end

  license "MIT"
  head "https://github.com/ccw/ccw.git", branch: "main"

  def install
    bin.install Dir["ccw*"].first => "ccw"
  end
end
