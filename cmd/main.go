// +build linux

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"syscall"
)

var (
	rootfs     *string
	volumeArg  *string
	cgroupName *string
	cleanup    *bool
)

func init() {
	rootfs = flag.String("rootfs", "", "the rootfs path")
	cgroupName = flag.String("cgroup", "", "the cgroup name")
	volumeArg = flag.String("volume", "", "the volume to mount")
	cleanup = flag.Bool("keep", false, "keep the rootfs on exit")
}

func main() {
	var cmd *exec.Cmd

	flag.Parse()
	args := flag.Args()

	if *rootfs == "" {
		panic("where is the rootfs?")
	}

	if args[0] == "child" {

		if *cgroupName != "" {
			must(addToCgroup(os.Getpid(), *cgroupName))
		}
		must(mountRootfs(*rootfs, *volumeArg))

		cmd = exec.Command(args[1], args[2:]...)
	} else {
		if *cgroupName != "" {
			must(createCgroup(*cgroupName))
		}
		cmd = exec.Command("/proc/self/exe", append([]string{"-rootfs", *rootfs, "-cgroup", *cgroupName, "-volume", *volumeArg, "child"}, args[0:]...)...)
		applyNamespaces(cmd)
	}

	forwardSignals(cmd)
	runCmd(cmd)

	if *cleanup {
		syscall.Unmount(*rootfs, 0)
		os.RemoveAll(*rootfs)
	}
	os.Exit(0)
}

func forwardSignals(cmd *exec.Cmd) {
	signalChannel := make(chan os.Signal)
	signal.Notify(signalChannel,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		for {
			select {
			case receivedSignal := <-signalChannel:
				err := cmd.Process.Signal(receivedSignal)
				if err != nil {
					fmt.Printf("sigkill failed: %s\n", err.Error())
				}
			}
		}
	}()
}

func must(err error) {
	if err != nil {
		fmt.Println("MUST FAILED")
		panic(err)
	}
}

func runCmd(cmd *exec.Cmd) {
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus := exitError.Sys().(syscall.WaitStatus)
			os.Exit(waitStatus.ExitStatus())
		}

		must(err)
	}
}

func createCgroup(name string) error {
	err := os.MkdirAll("/sys/fs/cgroup/cpuset/"+name, 0700)
	if err != nil {
		return err
	}

	err = os.MkdirAll("/sys/fs/cgroup/memory/"+name, 0700)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path.Join("/sys/fs/cgroup/cpuset", name, "cpuset.mems"), []byte("0"), 0700)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path.Join("/sys/fs/cgroup/cpuset", name, "cpuset.cpus"), []byte("0-3"), 0700)
	if err != nil {
		return err
	}

	return nil
}

func addToCgroup(pid int, cgroupName string) error {
	f, err := os.OpenFile(path.Join("/sys/fs/cgroup/cpuset", cgroupName, "cgroup.procs"), os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	pidStr := strconv.Itoa(pid)
	if _, err = f.WriteString(pidStr); err != nil {
		return err
	}

	f2, err := os.OpenFile(path.Join("/sys/fs/cgroup/memory", cgroupName, "cgroup.procs"), os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	defer f2.Close()

	if _, err = f2.WriteString(pidStr); err != nil {
		return err
	}

	return nil
}

func mountRootfs(rootfs string, volume string) error {
	must(syscall.Mount(rootfs, rootfs, "", syscall.MS_BIND, ""))
	must(os.MkdirAll(rootfs+"/oldrootfs", 0700))
	if volume != "" {
		volume := strings.Split(volume, ":")
		targetVolume := path.Join(rootfs, volume[1])
		must(os.MkdirAll(targetVolume, 0700))
		must(syscall.Mount(volume[0], targetVolume, "", syscall.MS_BIND, ""))
	}
	must(syscall.PivotRoot(rootfs, rootfs+"/oldrootfs"))
	must(os.Chdir("/"))
	must(syscall.Mount("", "/proc", "proc", 0, ""))
	return nil
}

func applyNamespaces(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}
}
