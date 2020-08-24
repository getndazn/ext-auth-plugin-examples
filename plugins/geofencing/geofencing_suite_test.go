package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGeoFencing(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GeoFencing Suite")
}
