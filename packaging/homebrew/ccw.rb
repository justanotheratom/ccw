class Ccw < Formula
  desc "Claude Code Workspace manager"
  homepage "https://github.com/ccw/ccw"
  version "0.1.10"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.10/ccw-darwin-arm64"
      sha256 "10b74ca944f39dc7594a5c123eb840db654b7737ca8d4a42767145b0a67e32b2"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.10/ccw-darwin-amd64"
      sha256 "921355d0c46a582570d3be36f7f9b6fc0bda87f1cb9dc3db9118df494fbaea18"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.10/ccw-linux-arm64"
      sha256 "eb82ffc28e48410fdf74a0f901170abb0452e074a09fd92f5698f7651b6a2b70"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.10/ccw-linux-amd64"
      sha256 "43c3deb11df5ab96e0621b0dc4345b077b3a9fef985acf1da3b9c767a73823e5"
    end
  end

  license "MIT"
  head "https://github.com/ccw/ccw.git", branch: "main"

  def install
    bin.install Dir["ccw*"].first => "ccw"
  end
end
