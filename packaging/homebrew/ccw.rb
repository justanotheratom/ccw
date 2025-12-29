class Ccw < Formula
  desc "Claude Code Workspace manager"
  homepage "https://github.com/ccw/ccw"
  version "0.1.17"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.17/ccw-darwin-arm64"
      sha256 "a0c9f4ba51f8a628d5150aba4c84f1646130e6214c031b9b358058fa32becb21"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.17/ccw-darwin-amd64"
      sha256 "9ee0e9e56a9b2309a9544a470c81b8a826c903273b7ef0c2c499ab6812c7c29a"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.17/ccw-linux-arm64"
      sha256 "88edb995f1a5db6aea9c7b42e2b449f656766953baf264c2007f6726e2e21264"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.17/ccw-linux-amd64"
      sha256 "751a92e69639e8d59fa7f700a49588fae1c4b6b46c9bb13222e4a91c0e0e5c5c"
    end
  end

  license "MIT"
  head "https://github.com/ccw/ccw.git", branch: "main"

  def install
    bin.install Dir["ccw*"].first => "ccw"
  end
end
