class Ccw < Formula
  desc "Claude Code Workspace manager"
  homepage "https://github.com/ccw/ccw"
  version "0.1.4"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.4/ccw-darwin-arm64"
      sha256 "52a84528bfa3df72b8a1d75c82a10a481936bf0a330691c5445c4be71152df0f"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.4/ccw-darwin-amd64"
      sha256 "72f4e3a541b95082055b605fbbda2516320fc415f80cc778170bffa15dce4e1c"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.4/ccw-linux-arm64"
      sha256 "4356b7455173198ac57952d30e01ef6d9f7ea4ee16a2af20ad522767670e0a7d"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.4/ccw-linux-amd64"
      sha256 "66d7dc59f337c08621c87c430782998fd6c7271611cac23cc8733151cd994976"
    end
  end

  license "MIT"
  head "https://github.com/ccw/ccw.git", branch: "main"

  def install
    bin.install Dir["ccw*"].first => "ccw"
  end
end
