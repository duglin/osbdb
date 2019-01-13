workflow "New workflow" {
  on = "push"
  resolves = ["Build"]
}

action "Build" {
  uses = "docker://golang"
  runs = "go get ... && /usr/bin/make"
}
