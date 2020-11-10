# all: svc seed  # ctrl+c out of this kills all the things...

help:
	egrep -o '^\w+[:]' $(MAKEFILE_LIST)

svc: nomad consul
	@echo
	@echo 'Nomad:  http://localhost:4646/ui/'
	@echo 'Consul: http://localhost:8500/ui/'
	@echo

logs:
	mkdir -p logs

nomad: | logs
	nomad agent -dev > logs/nomad.log &
	while true; do sleep 1 ; nomad node status 2>/dev/null && break ; printf '.'; done

consul: | logs
	consul agent -dev > logs/consul.log &
	while true; do sleep 0.5 ; consul members 2>/dev/null && break ; printf '.'; done

build:
	go build .

seed: build
	./hashicorp-game-of-life seed

ui:
	while true; do \
		./hashicorp-game-of-life ui 2>/dev/null ;\
		echo -------- ;\
		sleep 1 ;\
	done

killall:
	for y in $(shell seq 1 5); do \
	  for x in $(shell seq 1 5); do \
	    NOMAD_JOB_NAME=$$x-$$y ./hashicorp-game-of-life kill ;\
	  done \
	done

clean:
	nomad stop -purge 0-0 && sleep 10 || true
	nomad status | awk '/service/ {print$$1}' | while read j; do \
	  nomad stop -purge $$j || true ;\
	done

kill:
	pkill nomad consul hashicorp-game-of-life || true
	while true; do sleep 0.5; consul members    2>/dev/null || break ; done
	while true; do sleep 0.5; nomad node status 2>/dev/null || break ; done
	rm -rf ./logs/ /tmp/hgol/

.PHONY: all svc nomad consul seed clean
