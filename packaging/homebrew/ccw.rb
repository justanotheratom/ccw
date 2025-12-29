class Ccw < Formula
  desc "Claude Code Workspace manager"
  homepage "https://github.com/ccw/ccw"
  version "0.1.6"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.6/ccw-darwin-arm64"
      sha256 "1142013a325d0264b2d2239c1f3c1cda360b652bb508adc1f621eed49983223f"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.6/ccw-darwin-amd64"
      sha256 "e14268f7ce628b023f6df6c178111970dd900c16350d439e1c497e9633fc3f65"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.6/ccw-linux-arm64"
      sha256 "e86be6db7d07071c07bed226c033f5b03451c18b4dcba2985cc38a1a19121076"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.6/ccw-linux-amd64"
      sha256 "8eec7ae2d42f39c75ca87a967e1f3961bd4c90c996e726e4f912bddbb3b2d6cb"
    end
  end

  license "MIT"
  head "https://github.com/ccw/ccw.git", branch: "main"

  def install
    bin.install Dir["ccw*"].first => "ccw"
  end
end
