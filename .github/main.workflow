workflow "New workflow" {
  on = "push"
  resolves = ["Build"]
}

action "Build" {
  uses = "docker://golang"
  runs = "sh -c 'go get ... && /usr/bin/make'"
}
