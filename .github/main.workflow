workflow "New workflow" {
  on = "push"
  resolves = ["Build"]
}

action "Build" {
  uses = "docker://golang"
  runs = "/usr/bin/make"
}
