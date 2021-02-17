project = "hashicorp-game-of-life"

app "gol" {

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
      jobspec = file("${path.project}/seed.nomad")
    }
  }

}
