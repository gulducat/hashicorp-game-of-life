# HashiCorp's Game of Life

A distributed implementation of
[Conway’s Game of Life](https://en.wikipedia.org/wiki/Conway%27s_Game_of_Life)
using HashiCorp tools.

* Each “cell” is a separate process, scheduled by Nomad.

* They discover their neighbors’ addresses via Consul.

* Each apply rules to themselves, then report to neighbors and a “seed” job.

* All cells’ statuses are stored by the seed for us to view.

[HashiTalks presentation slides](https://docs.google.com/presentation/d/1VC7D6EYA2Z6ivHBX7RKJhc3ZyhEFFoCkAIA9ADt6U5A)
