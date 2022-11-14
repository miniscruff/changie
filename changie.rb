# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class Changie < Formula
  desc "Automated changelog tool for preparing releases with lots of customization options."
  homepage "https://changie.dev"
  version "1.10.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/miniscruff/changie/releases/download/v1.10.0/changie_1.10.0_darwin_arm64.tar.gz"
      sha256 "222f0747ac777242ba680e2551c6e880324e8bc3dafe00e6475c9c5901335b75"

      def install
        bin.install "changie"
      end
    end
    if Hardware::CPU.intel?
      url "https://github.com/miniscruff/changie/releases/download/v1.10.0/changie_1.10.0_darwin_amd64.tar.gz"
      sha256 "b6666d28cb4514dc139dbbae983abddf63086d968f5137f05e967cac0ef3afdc"

      def install
        bin.install "changie"
      end
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      url "https://github.com/miniscruff/changie/releases/download/v1.10.0/changie_1.10.0_linux_amd64.tar.gz"
      sha256 "2dc5743d567b00d31c0b37d0f2ada5808a5bdbbad887ae5771cdabfbd39ef90f"

      def install
        bin.install "changie"
      end
    end
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/miniscruff/changie/releases/download/v1.10.0/changie_1.10.0_linux_arm64.tar.gz"
      sha256 "b016bfbe8cbaefe60be6e434f665a163d0b6723b9c4a7a4385c6e99a3a27ba18"

      def install
        bin.install "changie"
      end
    end
  end
end
