export HOST ?= localhost
export NOMAD_ADDR ?= http://$(HOST):4646
export CONSUL_HTTP_ADDR ?= http://$(HOST):8500

all: svc seed

help:
	egrep -o '^\w+[:]' $(MAKEFILE_LIST)

svc: consul nomad
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
		curl http://$(HOST) ;\
		echo -------- ;\
		sleep 1 ;\
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
	@nomad status 0-0 | awk '/0-0.*running/ {print$$2}' | while read -r node; do \
	  nomad node status -verbose $$node | awk '/public-ipv4/ {print$$NF}' ;\
	  break ;\
	done

ui2:
	while true; do \
	  curl http://$(shell make get-ip 2>/dev/null) ;\
	  echo --- ;\
	done

clean:
	nomad stop -purge 0-0 && bash -c 'for x in {1..15}; do n="$$(nomad status | wc -l)"; echo $$x $$n; test $$n -le 2 && break; sleep 1; done' || true
	nomad status | awk '/system|service/ {print$$1}' | while read j; do \
		curl -sX DELETE $(NOMAD_ADDR)/v1/job/$$j?purge=true >/dev/null ;\
	done
	curl -X PUT $(NOMAD_ADDR)/v1/system/gc

kill:
	pkill nomad consul hashicorp-game-of-life || true
	while true; do sleep 0.5; consul members    2>/dev/null || break ; done
	while true; do sleep 0.5; nomad node status 2>/dev/null || break ; done
	rm -rf ./logs/ /tmp/hgol/

.PHONY: all svc nomad consul seed clean
