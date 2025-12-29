class Ccw < Formula
  desc "Claude Code Workspace manager"
  homepage "https://github.com/ccw/ccw"
  version "0.1.3"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.3/ccw-darwin-arm64"
      sha256 "340e67e3ced735be0f5ba211c27da1910f180d502d4fb6d17e48c81596adb7a6"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.3/ccw-darwin-amd64"
      sha256 "d431bd6d5ca4b872a889c947762a13ecb6551fda1082e6639c88826dccd6e46f"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.3/ccw-linux-arm64"
      sha256 "373a10eb0f3cab3477fea7ead6589c30e3afb3c7d6e579d6ba4ba056788a6e72"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.3/ccw-linux-amd64"
      sha256 "72b23941617efd0c2841302a1578c024d80741919ea944b817e8630f5e77ca8e"
    end
  end

  license "MIT"
  head "https://github.com/ccw/ccw.git", branch: "main"

  def install
    bin.install Dir["ccw*"].first => "ccw"
  end
end
