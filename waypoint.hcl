project = "hashicorp-game-of-life"

app "gol" {
  # config {
  #   env = {
  #     # CONSUL_HTTP_ADDR = "http://192.168.1.254:8500"
  #     # NOMAD_ADDR = "http://192.168.1.254:4646"
  #     NOMAD_JOB_NAME = "0-0"
  #   }
  # }

  build {
    use "docker" {}
    registry {
      use "docker" {
        image = "gulducat/hashicorp-game-of-life"
        tag   = "latest"
      }
    }
  }

  deploy {
    use "nomad" {
      jobspec = "seed.nomad"
      # jobspec = templatefile(
      #   "seed.nomad",
      #   {
      #     image = artifact.image
      #     another = env.WAYPOINT_WHATEVER
      #   }
      # )
    }
    # use "docker" {
    #   # command = ["-listen", ":80"]
    #   service_port = 80
    # }
    # use "exec" {
    #   # command = ["hashicorp-game-of-life", "seed"]
    #   command = ["nomad", "run", "seed.nomad"]
    # }
  }

  # TODO: make test happen?
  # test {

  # }
}
