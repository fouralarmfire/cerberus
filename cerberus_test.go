package cerberus_test

import (
  "os"
  "os/exec"

  . "github.com/onsi/ginkgo"
  . "github.com/onsi/gomega"
  . "github.com/onsi/gomega/gbytes"
  . "github.com/onsi/gomega/gexec"
)

var _ = Describe("Cerberus", func() {
  AfterEach(func() {
    os.Remove("/hello_world")
    os.Remove("/opt/busybox/hello_world")
  })

  It("runs stuff", func() {
    var buffer []byte
    stdout := BufferWithBytes(buffer)

    _, err := Start(exec.Command(cerberusPath, "-rootfs", "/opt/busybox", "echo", "Hello world"), stdout, stdout)
    Eventually(stdout).Should(Say("Hello world"))
    Expect(err).ToNot(HaveOccurred())
  })

  It("runs stuff in container", func() {
    var buffer []byte
    stdout := BufferWithBytes(buffer)

    _, err := Start(exec.Command(cerberusPath, "-rootfs", "/opt/busybox", "sh", "-c", "/bin/hostname foo; /bin/hostname"), stdout, stdout)
    Expect(err).ToNot(HaveOccurred())
    Eventually(stdout).Should(Say("foo"))
  })

  It("returns the exit code of a process running in a container", func() {
    var buffer []byte
    stdout := BufferWithBytes(buffer)

    session, err := Start(exec.Command(cerberusPath, "-rootfs", "/opt/busybox", "sh", "-c", "exit 12"), stdout, stdout)
    Expect(err).ToNot(HaveOccurred())
    Eventually(session.Wait()).Should(Exit(12))
  })

  It("runs processes inside an isolated root filesystem", func() {
    var buffer []byte
    stdout := BufferWithBytes(buffer)

    session, err := Start(exec.Command(cerberusPath, "-rootfs", "/opt/busybox", "sh", "-c", "touch /hello_world"), stdout, stdout)
    Expect(err).ToNot(HaveOccurred())
    Eventually(session.Wait()).Should(Exit(0))

    Expect("/hello_world").ToNot(BeAnExistingFile())
    Expect("/opt/busybox/hello_world").To(BeAnExistingFile())
  })
})
