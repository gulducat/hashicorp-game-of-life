job "api" {
  datacenters = ["dc1"]
  # type        = "system"
  task "api" {
    driver = "raw_exec"
    config {
      command = "hashicorp-game-of-life"
      args = ["api"]
    }
    service {
      name = "api"
    }
  }
}
