# all: services seed  # ctrl+c out of this kills all the things...

services: nomad consul

logs:
	mkdir -p logs

nomad: | logs
	nomad agent -dev > logs/nomad.log &
	while true; do sleep 1 ; nomad node status && break ; done

consul: | logs
	consul agent -dev > logs/consul.log &
	while true; do sleep 0.5 ; consul members && break ; done

build:
	go build .

seed: build
	./hashicorp-game-of-life run

grid:
	@# ./hashicorp-game-of-life grid 2>/dev/null
	./hashicorp-game-of-life grid

killall:
	for y in $(shell seq 1 5); do \
	  for x in $(shell seq 1 5); do \
	    NOMAD_JOB_NAME=$$x-$$y ./hashicorp-game-of-life kill ;\
	  done \
	done

clean:
	nomad status | awk '/service/ {print$$1}' | while read j; do \
	  nomad stop -purge $$j || true ;\
	done

kill:
	pkill nomad consul hashicorp-game-of-life   || true
	while true; do sleep 0.5; consul members    || break ; done
	while true; do sleep 0.5; nomad node status || break ; done
	rm -rf ./logs/

.PHONY: all services nomad consul seed clean
