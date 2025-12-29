class Ccw < Formula
  desc "Claude Code Workspace manager"
  homepage "https://github.com/ccw/ccw"
  version "0.1.21"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.21/ccw-darwin-arm64"
      sha256 "3f8da9883c05d40a1db4d5c27cb9b35363721d99eb74d8bc85b8b21a2b6764b0"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.21/ccw-darwin-amd64"
      sha256 "9bd7305ccb51b13c40adb44d330c1baf918722796a5515d4f9c8a1219266861a"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.21/ccw-linux-arm64"
      sha256 "53b174e1c311be0177fcb3890ff91828bf2fc6fb17eb4fa80f727b633a3d83cd"
    else
      url "https://github.com/justanotheratom/ccw/releases/download/v0.1.21/ccw-linux-amd64"
      sha256 "2cddbc64354490005e3cf74ef951b63d88494139641b02f935d6615429d90ac6"
    end
  end

  license "MIT"
  head "https://github.com/ccw/ccw.git", branch: "main"

  def install
    bin.install Dir["ccw*"].first => "ccw"
  end
end
