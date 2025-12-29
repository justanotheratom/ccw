class Ccw < Formula
  desc "Claude Code Workspace manager"
  homepage "https://github.com/ccw/ccw"
  version "0.1.19"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.19/ccw-darwin-arm64"
      sha256 "62c8e20888fa396a6ff6014cc2e6a82584fbf308ca392e9e85ef576be86ef41a"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.19/ccw-darwin-amd64"
      sha256 "b6ec3c169dc58de5af350050f538e19c2e935c05d863b6e1f0f120f3ebab54ec"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.19/ccw-linux-arm64"
      sha256 "dbc44d1d8075160ee588b09d6b65c2a223c819ebe203151b85fa8dd4181b6c16"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.19/ccw-linux-amd64"
      sha256 "4dcea0c20ceb3fc74f83a1b9e83ed634c498724de38d49afa6032a3f2788c27d"
    end
  end

  license "MIT"
  head "https://github.com/ccw/ccw.git", branch: "main"

  def install
    bin.install Dir["ccw*"].first => "ccw"
  end
end
