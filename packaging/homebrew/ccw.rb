class Ccw < Formula
  desc "Claude Code Workspace manager"
  homepage "https://github.com/ccw/ccw"
  version "0.1.13"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.13/ccw-darwin-arm64"
      sha256 "555278bbd7ae1389d0815c022ce585bc90508c42af5a00f867f0647c4b85ffd3"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.13/ccw-darwin-amd64"
      sha256 "9b91c4bccfa800430e270d7bc172a72214ed6918e4b310022c019779ceaf22ec"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.13/ccw-linux-arm64"
      sha256 "748701285e77cbec33ef3166b5b6a1d1750a8f8afb27556f1b079387fb03ca56"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.13/ccw-linux-amd64"
      sha256 "2eae54954bb7cae42289689f951b376a2078af315563b964d62471f368c44354"
    end
  end

  license "MIT"
  head "https://github.com/ccw/ccw.git", branch: "main"

  def install
    bin.install Dir["ccw*"].first => "ccw"
  end
end
