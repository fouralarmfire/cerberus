package main

import (
  "flag"
  "io/ioutil"
  "path"
)

var (
  cgroupName *string
  memLimit   *string
)

func init() {
  cgroupName = flag.String("cgroup", "", "the cgroup name")
  memLimit = flag.String("max", "", "the max memory")
}

func main() {
  flag.Parse()

  err := ioutil.WriteFile(path.Join("/sys/fs/cgroup/memory", *cgroupName, "memory.limit_in_bytes"), []byte(*memLimit), 0700)
  if err != nil {
    panic(err)
  }
}
