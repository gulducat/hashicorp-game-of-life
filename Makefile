export HOST ?= localhost
export NOMAD_ADDR ?= http://$(HOST):4646
export CONSUL_HTTP_ADDR ?= http://$(HOST):8500

all: svc seed

help:
	egrep -o '^\w+[:]' $(MAKEFILE_LIST)

svc: nomad consul
	@echo
	@echo 'Nomad:  $(NOMAD_ADDR)/ui/'
	@echo 'Consul: $(CONSUL_HTTP_ADDR)/ui/'
	@echo

logs:
	mkdir -p logs

nomad: | logs
	nomad agent -dev > logs/nomad.log &
	while true; do sleep 1 ; nomad node status 2>/dev/null && break ; printf '.'; done

consul: | logs
	consul agent -dev > logs/consul.log &
	# consul agent -dev >/dev/null &
	while true; do sleep 0.5 ; consul members 2>/dev/null && break ; printf '.'; done

build:
	go build .
	cp hashicorp-game-of-life /usr/local/bin/

api:
	nomad run api.nomad

seed: build
	./hashicorp-game-of-life seed

ui:
	while true; do \
		curl -s http://$(HOST) ;\
		echo -------- ;\
		sleep 0.5 ;\
	done

more:
	./hashicorp-game-of-life more

tail:
	tail -f logs/*

s3:
	aws s3 cp hashicorp-game-of-life s3://game-of-life-hackathon/hashicorp-game-of-life

upload:
	GOARCH=amd64 GOOS=linux go build -o gol-linux .
	for l in $(shell cat servers.list); do \
	  rsync -avP gol-linux $$l:~/ ;\
	  ssh $$l 'sudo cp gol-linux /usr/local/bin/hashicorp-game-of-life' ;\
	done

get-ip:
	@nomad status api | awk '/running/ {print$$2}' | while read -r node; do \
	  nomad node status -verbose $$node | awk '/public-ipv4/ {print$$NF}' ;\
	done

ui2:
	while true; do \
	  curl -s http://$(shell make get-ip 2>/dev/null) ;\
	  echo --- ;\
	done

clean:
	nomad stop -purge 0-0 && sleep 5 || true
	nomad status | awk '/service/ {print$$1}' | while read j; do \
	  nomad stop -purge $$j || true ;\
	done

kill:
	pkill nomad consul hashicorp-game-of-life || true
	while true; do sleep 0.5; consul members    2>/dev/null || break ; done
	while true; do sleep 0.5; nomad node status 2>/dev/null || break ; done
	rm -rf ./logs/ /tmp/hgol/

.PHONY: all svc nomad consul seed clean
