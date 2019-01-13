workflow "New workflow" {
  on = "push"
  resolves = ["Build"]
}

action "Build" {
  uses = "docker://golang"
  runs = ["/bin/sh"]
  args = ["-c","go get -d . && /usr/bin/make"]
}
