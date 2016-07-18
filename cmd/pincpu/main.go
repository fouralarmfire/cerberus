package main

import (
  "flag"
  "io/ioutil"
  "path"
)

var (
  cgroupName *string
  cpusNumber *string
)

func init() {
  cgroupName = flag.String("cgroup", "", "the cgroup name")
  cpusNumber = flag.String("cpus", "", "the cgroup name")
}

func main() {
  flag.Parse()

  err := ioutil.WriteFile(path.Join("/sys/fs/cgroup/cpuset", *cgroupName, "cpuset.cpus"), []byte(*cpusNumber), 0700)
  if err != nil {
    panic(err)
  }
}
