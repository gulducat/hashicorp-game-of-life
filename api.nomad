job "api" {
  datacenters = ["dc1"]
  # type        = "system"
  priority    = 90
  group "api" {
    network {
      port "udp" {}
    }
    restart {
      attempts = 5
      delay    = "3s"  # 3 000000000
    }
    task "api" {
      driver = "raw_exec"
      config {
        command = "hashicorp-game-of-life"
        args = ["api"]
      }
      service {
        name = "api"
        port = "udp"
      }
      resources {
        cpu    = 1600
        memory = 60
      }
    }
  }
}
