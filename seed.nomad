locals {
  # good for 8 clients (c5.2xlarge), ~270 per node
  w = 56
  h = 38
  # good for laptop
  # w = 9
  # h = 7
}

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
        static = 8080
      }
    }
    service {
      name = "0-0"
      port = "udp"
    }
    service {
      name = "0-0-http"
      port = "http"
    }
    task "0-0" {
      driver = "raw_exec"
      config {
        # command = "/Users/danielbennett/git/gulducat/hashicorp-game-of-life/mathy/mathy"
        command = "hashicorp-game-of-life"
        args    = ["run"]
        # command = "bash"
        # args    = ["-c", "env | grep NAME; sleep 3600"]
      }
      env {
        MAX_W = local.w
        MAX_H = local.h
      }
      resources {
        cpu    = 1200
        memory = 100
      }
    }
  }

  group "cells" {
    count = (local.w * local.h) + 1
    restart {
      attempts = 10
      delay    = "5s"
    }
    network {
      port "udp" {}
    }
    service {
      name = "CELL-${NOMAD_ALLOC_INDEX}"
      port = "udp"
    }
    task "cell" {
      driver = "raw_exec"
      config {
        # command = "/Users/danielbennett/git/gulducat/hashicorp-game-of-life/mathy/mathy"
        command = "hashicorp-game-of-life"
        args    = ["run"]
      }
      env {
        MAX_W = local.w
        MAX_H = local.h
      }
      resources {
        # each job doesn't really need this much cpu, but things go sideways below this value (on laptop...?)
        # and this gets us to ~180 jobs per client anyway, which is pretty solid.
        cpu    = 100
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
