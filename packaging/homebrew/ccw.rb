class Ccw < Formula
  desc "Claude Code Workspace manager"
  homepage "https://github.com/ccw/ccw"
  version "0.1.2"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.2/ccw-darwin-arm64"
      sha256 "5d11a1f25bc3edc0573eea4236bede8ddc0b9dd94dec9aa79eb2840b39d0f216"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.2/ccw-darwin-amd64"
      sha256 "26b0b74e338aced6c135188aad59ce984152a2b3e8dba04c885312779e57e378"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.2/ccw-linux-arm64"
      sha256 "ffc809bdc2c855842441e6174eeae5dc4e1534336f45ff270c505744dec53667"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.2/ccw-linux-amd64"
      sha256 "21dabc0ed82b1cd5aceca0060cbbbc18979f317896ed75f297a693e70e9995e5"
    end
  end

  license "MIT"
  head "https://github.com/ccw/ccw.git", branch: "main"

  def install
    bin.install Dir["ccw*"].first => "ccw"
  end
end
