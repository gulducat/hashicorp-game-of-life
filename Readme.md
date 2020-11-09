# HashiCorp's Game of Life

[Conway’s Game of Life](https://en.wikipedia.org/wiki/Conway%27s_Game_of_Life)
where each cell is a Nomad job and “alive/dead” is the
health check status of each job’s registered Consul service.

Following Conway’s rules, each cell checks its neighbors’ health
via Consul to determine what its own health should be.
