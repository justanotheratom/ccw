class Ccw < Formula
  desc "Claude Code Workspace manager"
  homepage "https://github.com/ccw/ccw"
  version "0.1.5"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.5/ccw-darwin-arm64"
      sha256 "563a7ec6c1b3355eca0be8f7a6b9627122aabd7d660693e0ba75c65d735d717c"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.5/ccw-darwin-amd64"
      sha256 "33b3814c21ad7ae9f8e9709697b66f996bbc3189a6b3a01912e82e40aee5d4ca"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.5/ccw-linux-arm64"
      sha256 "e86be6db7d07071c07bed226c033f5b03451c18b4dcba2985cc38a1a19121076"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.5/ccw-linux-amd64"
      sha256 "4450733268de3545fdc7480c657f5e47f5eaa5427431ae6ba0138e8068a08e33"
    end
  end

  license "MIT"
  head "https://github.com/ccw/ccw.git", branch: "main"

  def install
    bin.install Dir["ccw*"].first => "ccw"
  end
end
