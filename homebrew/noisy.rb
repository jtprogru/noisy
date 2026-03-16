class Noisy < Formula
  desc "Random HTTP/DNS traffic noise generator"
  homepage "https://github.com/jtprogru/noisy"
  version "0.1.0"

  if OS.mac?
    url "https://github.com/jtprogru/noisy/archive/refs/tags/v#{version}.tar.gz"
    sha256 "ae3dacbba037353b6ff32f96e6608cf3f23e0730482c26f9bba31d37f8fea43e"
  elsif OS.linux?
    url "https://github.com/jtprogru/noisy/archive/refs/tags/v#{version}.tar.gz"
    sha256 "ae3dacbba037353b6ff32f96e6608cf3f23e0730482c26f9bba31d37f8fea43e"
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
