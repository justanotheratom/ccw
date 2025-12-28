class Ccw < Formula
  desc "Claude Code Workspace manager"
  homepage "https://github.com/ccw/ccw"
  version "0.1.0"

  url "https://github.com/ccw/ccw/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "ADD_SHA256_HERE"
  license "MIT"
  head "https://github.com/ccw/ccw.git", branch: "main"

  depends_on "go" => :build
  depends_on "tmux"

  def install
    system "go", "build", *std_go_args, "./..."
  end
end
