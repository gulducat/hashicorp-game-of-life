export HOST ?= $(shell ifconfig | awk '/192\./ {print$$2}')
export NOMAD_ADDR ?= http://$(HOST):4646
export CONSUL_HTTP_ADDR ?= http://$(HOST):8500
NETWORK ?= en0

.PHONY: env all help svc nomad consul build api seed ui more tail s3 upload get-ip ui2 clean kill

env:
	@echo 'export NOMAD_ADDR=$(NOMAD_ADDR) && export CONSUL_HTTP_ADDR=$(CONSUL_HTTP_ADDR)'

all: svc seed

help:
	egrep -o '^\w+[:]' $(MAKEFILE_LIST)

svc: consul nomad
	@echo
	@echo export NOMAD_ADDR=$(NOMAD_ADDR)
	@echo export CONSUL_HTTP_ADDR=$(CONSUL_HTTP_ADDR)
	@echo

logs:
	mkdir -p logs

nomad: | logs
	nomad agent -dev -bind=$(HOST) -network-interface=$(NETWORK) > logs/nomad.log &
	while true; do sleep 1 ; nomad node status 2>/dev/null && break ; printf '.'; done

consul: | logs
	consul agent -dev -bind=$(HOST) -advertise=$(HOST) -client=$(HOST) > logs/consul.log &
	# consul agent -dev >/dev/null &
	while true; do sleep 0.5 ; consul members 2>/dev/null && break ; printf '.'; done

build:
	go build .
	cp hashicorp-game-of-life /usr/local/bin/
	# docker build -t gol:local .

p-%:
	hashicorp-game-of-life pattern $*

api:
	nomad run api.nomad

seed: #build
	nomad run seed.nomad

ui:
	while true; do \
		curl http://$(HOST) ;\
		echo -------- ;\
		sleep 0.5 ;\
	done

more:
	./hashicorp-game-of-life more

tail:
	tail -f logs/*

s3:
	aws s3 cp hashicorp-game-of-life s3://game-of-life-hackathon/hashicorp-game-of-life

upload: servers.list
	GOARCH=amd64 GOOS=linux go build -o gol-linux .
	for l in $(shell cat servers.list); do \
	  rsync -avP gol-linux ubuntu@$$l:~/ ;\
	  ssh ubuntu@$$l 'sudo cp gol-linux /usr/local/bin/hashicorp-game-of-life' ;\
	done

get-ip:
	@echo $(shell curl -sS $(CONSUL_HTTP_ADDR)/v1/catalog/service/0-0-http | jq -r '.[].Address'):$(shell curl -sS $(CONSUL_HTTP_ADDR)/v1/catalog/service/0-0-http | jq -r '.[].ServicePort')

# get-ip:
# 	@echo localhost:$(shell curl -s $(CONSUL_HTTP_ADDR)/v1/catalog/service/0-0-http | jq -r '.[].ServicePort')

ui2:
	while true; do \
	  curl http://$(shell make get-ip 2>/dev/null) ;\
	  echo --- ;\
	  sleep 0.5 ;\
	done

clean:
	#nomad stop -purge 0-0 && bash -c 'for x in {1..15}; do n="$$(nomad status | wc -l)"; echo $$x $$n; test $$n -le 2 && break; sleep 1; done' || true
	nomad status | awk '/system|service/ {print$$1}' | while read j; do \
		curl -sX DELETE $(NOMAD_ADDR)/v1/job/$$j?purge=true >/dev/null ;\
	done
	curl -X PUT $(NOMAD_ADDR)/v1/system/gc

kill:
	pkill nomad || true
	pkill consul || true
	ps aux | awk '/hashicorp-gam[e]/ {print$$2}' | xargs kill || true
	while true; do sleep 0.5; consul members    2>/dev/null || break ; done
	while true; do sleep 0.5; nomad node status 2>/dev/null || break ; done
	rm -rf ./logs/ /tmp/hgol/

.PHONY: all svc nomad consul seed clean

servers.list:
	aws --region us-east-1 \
	  ec2 describe-instances \
	  --filters 'Name=tag-value,Values=dbennett-nomad' \
	  | jq -r '.Reservations[].Instances[].PublicIpAddress' > servers.list


.PHONY: wp
wp:
	# env | grep -E '()' | while read env; do waypoint config set "$$env"; done
	waypoint config set -app gol NOMAD_ADDR=$(NOMAD_ADDR) CONSUL_HTTP_ADDR=$(CONSUL_HTTP_ADDR)
