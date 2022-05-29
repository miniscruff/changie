package cmd

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
)

func TestCmd(t *testing.T) {
	format.TruncatedDiff = false
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cmd Suite")
}
