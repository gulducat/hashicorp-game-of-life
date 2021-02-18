export SERVER ?= $(shell ifconfig | awk '/192\./ {print$$2}')
export NOMAD_ADDR ?= http://$(SERVER):4646
export CONSUL_HTTP_ADDR ?= http://$(SERVER):8500
NETWORK ?= en0
URL ?= http://$(SERVER)

.PHONY: help env

help:
	egrep -o '^\w+[:]' $(MAKEFILE_LIST)

env:
	@echo 'export NOMAD_ADDR=$(NOMAD_ADDR) && export CONSUL_HTTP_ADDR=$(CONSUL_HTTP_ADDR)'

## app ##

.PHONY: build run ui web p-% clean

build:
	go build .
	cp hashicorp-game-of-life /usr/local/bin/
	# docker build -t gol:local .

run: servers.map
	@echo nomad run traefik.nomad
	@echo nomad run gol.nomad
	@nomad run -var="consul_http_addr=http://$(shell awk '/$(SERVER)/ {print$$2}' servers.map):8500" traefik.nomad >/dev/null
	@nomad run -var="consul_http_addr=http://$(shell awk '/$(SERVER)/ {print$$2}' servers.map):8500" gol.nomad >/dev/null

ui:
	while true; do \
		curl $(URL)/raw ;\
		echo -------- ;\
		sleep 0.1 ;\
	done

link: servers.map
	@echo $(URL)

p-%:
	curl -sS $(URL)/p/$* | grep -v html
	@echo $(URL)

clean:
	nomad stop -purge traefik || true
	nomad stop -purge gol || true
	curl -X PUT $(NOMAD_ADDR)/v1/system/gc

## infra ##

.PHONY: infra-up infra-down ping upload

infra-up:
	cd terraform && terraform init && terraform apply -auto-approve

infra-down: clean
	cd terraform && terraform destroy -auto-approve

servers.map:
	aws --region us-east-1 \
	  ec2 describe-instances \
	  --filters 'Name=tag-value,Values=dbennett-nomad' \
	  | jq -r '.Reservations[].Instances[] | "\(.PublicIpAddress) \(.PrivateIpAddress)"' \
	  | grep -v null \
	  > servers.map

ping: servers.map
	awk '{print$$1}' servers.map | while read -r s; do \
	  printf "$$s " ;\
	  ping -c1 -t1 $$s >/dev/null && echo ✅ || echo 💔 ;\
	done

upload: servers.map | ping
	GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -o gol-linux .
	for l in $(shell awk '{print $$1}' servers.map); do \
	  rsync -avP gol-linux ubuntu@$$l:~/ >/dev/null || exit 1 ;\
	  ssh ubuntu@$$l 'sudo cp gol-linux /usr/local/bin/hashicorp-game-of-life' || exit 2 ;\
	  printf '.' ;\
	done
	@echo

## local dev ##

.PHONY: local nomad consul logs kill

local: consul nomad
	@echo
	@echo export NOMAD_ADDR=$(NOMAD_ADDR)
	@echo export CONSUL_HTTP_ADDR=$(CONSUL_HTTP_ADDR)
	@echo
	@echo $(SERVER) $(SERVER) > servers.map

nomad:
	@mkdir -p logs
	nomad agent -dev -bind=$(SERVER) -network-interface=$(NETWORK) > logs/nomad.log &
	while true; do sleep 1 ; nomad node status 2>/dev/null && break ; printf '.'; done

consul:
	@mkdir -p logs
	consul agent -dev -bind=$(SERVER) -advertise=$(SERVER) -client=$(SERVER) > logs/consul.log &
	while true; do sleep 0.5 ; consul members 2>/dev/null && break ; printf '.'; done

logs:
	tail -f logs/*

kill:
	pkill nomad || true
	pkill consul || true
	ps aux | awk '/hashicorp-gam[e]/ {print$$2}' | xargs kill || true
	while true; do sleep 0.5; consul members    2>/dev/null || break ; done
	while true; do sleep 0.5; nomad node status 2>/dev/null || break ; done
	rm -rf ./servers.* ./logs/ /tmp/hgol/


## DEMO ##

.PHONY: demo seed-alloc seed.log

# demo: run seed.log

seed-alloc:
	@for x in {1..5}; do \
	  nomad status gol | awk '/seed.*runn/ {print $$1}' 2>/dev/null | grep . && break ;\
	  sleep 1 ;\
	done

seed.log:
	@rm -f seed.log
	nomad logs -tail -f -stderr $(shell $(MAKE) seed-alloc) | cut -d' ' -f3- | tee seed.log
