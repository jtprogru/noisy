class Noisy < Formula
  desc "Random HTTP/DNS traffic noise generator"
  homepage "https://github.com/jtprogru/noisy"
  version "0.1.0"

  if OS.mac?
    url "https://github.com/jtprogru/noisy/archive/v#{version}.tar.gz"
    sha256 "CHANGE_ME" # Run `shasum -a 256 noisy-0.1.0.tar.gz` after creating release
  elsif OS.linux?
    url "https://github.com/jtprogru/noisy/archive/v#{version}.tar.gz"
    sha256 "CHANGE_ME"
  end

  license "GPL-3.0"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args, "-ldflags", "-s -w"
  end

  test do
    output = shell_output("#{bin}/noisy --help")
    assert_match "config file path", output
  end
end
