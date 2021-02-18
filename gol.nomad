locals {
  # good for 7 clients (c5.2xlarge), ~300 per node
  # w = 58
  # h = 36
  # good for laptop
  w = 4
  h = 4

  http = 8081
}

variable "consul_http_addr" {}

job "gol" {
  datacenters = ["dc1"]

  group "seed" {
    restart {
      attempts = 10
      delay    = "1s"
    }
    network {
      port "udp" {}
      port "http" {
        static = local.http
      }
    }
    service {
      name = "0-0"
      port = "udp"
    }
    service {
      name = "0-0-http"
      port = "http"
      tags = [
        "traefik.enable=true",
        "traefik.http.routers.seed.tls=false",
        "traefik.http.routers.seed.rule=Path(`/`) || Path(`/raw`) || PathPrefix(`/p`)",
      ]
    }
    task "seed" {
      driver = "raw_exec"
      config {
        command = "hashicorp-game-of-life"
        args    = ["run"]
      }
      # leaving docker here from trying to get waypoint to work
      # driver = "docker"
      # config {
      #   # image = "gol:local"
      #   image = "gulducat/hashicorp-game-of-life:latest"
      #   # image = "${image}"
      #   ports = ["http", "udp"]
      # }
      env {
        # PORT             = local.http # for waypoint
        MAX_W            = local.w
        MAX_H            = local.h
        CONSUL_HTTP_ADDR = var.consul_http_addr
      }
      resources {
        cpu    = 1200
        memory = 100
      }
    }
  }

  group "grid" {
    count = local.w * local.h + 1
    restart {
      attempts = 10
      delay    = "5s"
    }
    network {
      port "udp" {}
    }
    service {
      name = "cell-${NOMAD_ALLOC_INDEX}"
      port = "udp"
    }
    task "cell" {
      driver = "raw_exec"
      config {
        command = "hashicorp-game-of-life"
        args    = ["run"]
      }
      env {
        MAX_W            = local.w
        MAX_H            = local.h
        CONSUL_HTTP_ADDR = var.consul_http_addr
      }
      resources {
        # each job doesn't really need this much cpu, but things go sideways below this value
        cpu    = 90
        memory = 50
      }
      logs {
        max_files     = 1
        max_file_size = 10
      }
    }
    ephemeral_disk {
      size = 20
    }
  }
}
