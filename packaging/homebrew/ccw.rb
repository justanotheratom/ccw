class Ccw < Formula
  desc "Claude Code Workspace manager"
  homepage "https://github.com/ccw/ccw"
  version "0.1.9"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.9/ccw-darwin-arm64"
      sha256 "098eb468651611eec1af32089488ea8a8b9e990e74554cb7449d7bdb72ba0272"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.9/ccw-darwin-amd64"
      sha256 "099e9747d55e801169870550265ac13a8d99f7c22653e677daa6b80d70f9508a"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.9/ccw-linux-arm64"
      sha256 "f5767071359b1c327b0342127ec9dc701e7659e8c24e25da791c74787da2c4af"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.9/ccw-linux-amd64"
      sha256 "975848b6baf07b837dd05ec80824b0509cc925cfebb01618cbc6e353b14bb1d9"
    end
  end

  license "MIT"
  head "https://github.com/ccw/ccw.git", branch: "main"

  def install
    bin.install Dir["ccw*"].first => "ccw"
  end
end
