class Ccw < Formula
  desc "Claude Code Workspace manager"
  homepage "https://github.com/ccw/ccw"
  version "0.1.16"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.16/ccw-darwin-arm64"
      sha256 "7fab1fd0ab57bff0cb38274b587a00367c388fcb5be5b50ad5725afe1b64a398"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.16/ccw-darwin-amd64"
      sha256 "4d247ab23ae12eb2211b78a7935f811765f037482debac22b99427e097c44a2b"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.16/ccw-linux-arm64"
      sha256 "040ddf97ffac85744704321eff1a5b8a603f3a927798850cf35a3ba695c9d963"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.16/ccw-linux-amd64"
      sha256 "8e3a985cac9dcef14364ee8bff756dca525fdf9065556b745c842f12bb5977a4"
    end
  end

  license "MIT"
  head "https://github.com/ccw/ccw.git", branch: "main"

  def install
    bin.install Dir["ccw*"].first => "ccw"
  end
end
