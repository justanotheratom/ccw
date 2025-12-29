class Ccw < Formula
  desc "Claude Code Workspace manager"
  homepage "https://github.com/ccw/ccw"
  version "0.1.7"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.7/ccw-darwin-arm64"
      sha256 "feb04253ddc45c241f6dd5718b5023ad5b82a8259133b9e653dc2e7bcd03125d"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.7/ccw-darwin-amd64"
      sha256 "6b6f043baa19ea424365ab0217a8a897fa4685d977bb2f48c331b259138f9e75"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.7/ccw-linux-arm64"
      sha256 "e86be6db7d07071c07bed226c033f5b03451c18b4dcba2985cc38a1a19121076"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.7/ccw-linux-amd64"
      sha256 "80b6defce7ce943040e9bb6bcc26087f1aee0fd87fe56a1ed51b89c430913764"
    end
  end

  license "MIT"
  head "https://github.com/ccw/ccw.git", branch: "main"

  def install
    bin.install Dir["ccw*"].first => "ccw"
  end
end
