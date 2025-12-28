class Ccw < Formula
  desc "Claude Code Workspace manager"
  homepage "https://github.com/ccw/ccw"
  version "0.1.0"

  url "https://github.com/justanotheratom/ccw/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "a2cf5f47b87ed51615bc544f8e2403e05d4203a66046bd1800144f7b5f5fa8c7"
  license "MIT"
  head "https://github.com/ccw/ccw.git", branch: "main"

  depends_on "go" => :build
  depends_on "tmux"

  def install
    system "go", "build", *std_go_args, "./..."
  end
end
