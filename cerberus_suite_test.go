package cerberus_test

import (
	"fmt"
	"os"
	"runtime"

	"github.com/fatih/color"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

var cerberusPath string

var _ = BeforeSuite(func() {
	var err error
	if runtime.GOOS != "linux" {
		fmt.Println(color.RedString("will only run in linux"))
		os.Exit(0)
	}
	cerberusPath, err = gexec.Build("./cmd")
	Expect(err).ToNot(HaveOccurred())
})

func TestCerberus(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cerberus Suite")
}
