workflow "New workflow" {
  on = "push"
  resolves = ["Build"]
}

action "Build" {
  uses = "docker://golang"
  args = "[\"-c\",\"go get ... && /usr/bin/make\"]"
  runs = "[\"/bin/sh\"]"
}
