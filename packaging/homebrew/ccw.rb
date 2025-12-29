class Ccw < Formula
  desc "Claude Code Workspace manager"
  homepage "https://github.com/ccw/ccw"
  version "0.1.12"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.12/ccw-darwin-arm64"
      sha256 "5df9484ca0032d0814ad6df8c2c89d84335b89e5196f376ef5fd1100806fdd6a"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.12/ccw-darwin-amd64"
      sha256 "267f0ab089009efdd6f2b0a7ebf68380fba5b4a0613b75b5b9db50b4c0a0235c"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.12/ccw-linux-arm64"
      sha256 "e39806fd50bbd33d5a51f3d41ef707aec0c8aa6e717300e9759c5a810aba5456"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.12/ccw-linux-amd64"
      sha256 "aa1cced07afc193ac86c7f1ec4cd8bc4f53258e3ee45078f3254007272dac82c"
    end
  end

  license "MIT"
  head "https://github.com/ccw/ccw.git", branch: "main"

  def install
    bin.install Dir["ccw*"].first => "ccw"
  end
end
