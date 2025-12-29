class Ccw < Formula
  desc "Claude Code Workspace manager"
  homepage "https://github.com/ccw/ccw"
  version "0.1.11"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.11/ccw-darwin-arm64"
      sha256 "7e465b897c96ef9dd5e185de0bcc6e80728185ecb5d8a569d72e98f787b9f272"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.11/ccw-darwin-amd64"
      sha256 "17e1bdf79e94ae5dae2e002e2d7f0362939ab1d722b2a2594ccf08e839369069"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.11/ccw-linux-arm64"
      sha256 "bb3c60fff6420cb3d7a223f95f7c2f88cfc0b934b5f8e56f0b480289f1aba9ca"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.11/ccw-linux-amd64"
      sha256 "93a11a0fa19419944438d21b6b527b258775720966220ee822c1377d04aa5223"
    end
  end

  license "MIT"
  head "https://github.com/ccw/ccw.git", branch: "main"

  def install
    bin.install Dir["ccw*"].first => "ccw"
  end
end
