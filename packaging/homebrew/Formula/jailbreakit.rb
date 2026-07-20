class Jailbreakit < Formula
  desc "iOS pentest lab setup helper for authorized security testing"
  homepage "https://github.com/Waariss/jailbreakit"
  url "https://github.com/Waariss/jailbreakit/archive/refs/tags/v1.4.1.tar.gz"
  sha256 "8f7ea836f42eb6795e563c217dfbd688c394e71fc30cb03aff04ea7fda585cd2"
  license "MIT"

  depends_on "go" => :build

  def install
    ldflags = %W[
      -s
      -w
      -X github.com/Waariss/jailbreakit/internal/version.Version=v#{version}
      -X github.com/Waariss/jailbreakit/internal/version.Author=Waariss
    ]
    system "go", "build", "-trimpath", "-ldflags", ldflags.join(" "), "-o", bin/"jailbreakit", "./cmd/jailbreakit"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/jailbreakit version")
    assert_match "iOS pentest lab setup helper", shell_output("#{bin}/jailbreakit --help")
  end
end
