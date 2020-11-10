# HashiCorp's Game of Life

[Conway’s Game of Life](https://en.wikipedia.org/wiki/Conway%27s_Game_of_Life)
where each cell is a Nomad job and “alive/dead” is the
health check status of each job’s registered Consul service.

Following Conway’s rules, each cell checks its neighbors’ health
via Consul to determine what its own health should be.

## How Do

Pre-requisites: Go, Nomad, and Consul installed on your machine.

```shell
make svc    # run nomad and consul
make seed   # build and start "seed" job ("0-0")
make ui     # display terminal UI
make clean  # stop all nomad jobs
make kill   # kill nomad, consul, and hashicorp-game-of-life
```
