job "0-0" {
  datacenters = ["dc1"]

  group "cell" {

    network {
      port "udp" {}
      port "http" {
        to = 8080
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

    task "cell" {
      driver = "raw_exec"
      config {
        command = "hashicorp-game-of-life"
        args    = ["run"]
      }
      # driver = "docker"
      # config {
      #   image = "${tha_image_yo}"
      #   ports = ["udp", "http"]
      # }
      env {
        CONSUL_HTTP_ADDR = "http://localhost:8500"
        # NOMAD_ADDR = "http://192.168.1.254:4646"
        # CONSUL_HTTP_ADDR = "http://192.168.1.254:8500"
        # NOMAD_ADDR = "http://192.168.1.254:4646"
        # # waypoint entrypoint needs this
        PORT = 8080
      }
      resources {
        cpu    = 160
        memory = 35
      }
      restart {
        attempts = 10
        delay    = "1s"
      }
    }

  }

}
