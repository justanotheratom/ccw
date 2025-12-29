class Ccw < Formula
  desc "Claude Code Workspace manager"
  homepage "https://github.com/ccw/ccw"
  version "0.1.18"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.18/ccw-darwin-arm64"
      sha256 "003b0050c18d171abe9e0a86f9f71a19a9c1d1b9d3736caf488d1fb3b8648350"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.18/ccw-darwin-amd64"
      sha256 "b78b4cbd05dbb708ed9179a2f173a997fd1910b198f96a4b9dc44cd0ff06254b"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.18/ccw-linux-arm64"
      sha256 "aa15e2ee7b9d705614940fc40892db11226768961679505d0e908e457ea89ff1"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.18/ccw-linux-amd64"
      sha256 "a899b753d77af3e88eaa44cac823826ada1c51d0062236a6ca5828a2556e619f"
    end
  end

  license "MIT"
  head "https://github.com/ccw/ccw.git", branch: "main"

  def install
    bin.install Dir["ccw*"].first => "ccw"
  end
end
