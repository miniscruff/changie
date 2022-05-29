package core

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
)

func TestInternal(t *testing.T) {
	format.TruncatedDiff = false
	RegisterFailHandler(Fail)
	RunSpecs(t, "Core Suite")
}
