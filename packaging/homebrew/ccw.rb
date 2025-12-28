class Ccw < Formula
  desc "Claude Code Workspace manager"
  homepage "https://github.com/ccw/ccw"
  version "0.1.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.0/ccw-darwin-arm64"
      sha256 "fe02359567de878f5ad550967fd6cb0126657d157df85b9d002dd184de457338"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.0/ccw-darwin-amd64"
      sha256 "87f49d3645ca10b63fe7fac85a101e725608854f4aeb40919bd8eb7b0034187b"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.0/ccw-linux-arm64"
      sha256 "18eafe2b4b4292b5c5c340e4997a45a61bce8c178f990cede6b54ec04f9207f5"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.0/ccw-linux-amd64"
      sha256 "82a7eee3c774e121fb72a8f889f76b339efb48d5327a372ef903cce2a505a407"
    end
  end

  license "MIT"
  head "https://github.com/ccw/ccw.git", branch: "main"

  def install
    bin.install Dir["ccw*"].first => "ccw"
  end
end
