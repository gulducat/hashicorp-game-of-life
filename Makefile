services: nomad consul

logs:
	mkdir -p logs

nomad: | logs
	nomad agent -dev > logs/nomad.log &
	while true; do sleep 0.5 ; nomad node status && break ; done

consul: | logs
	consul agent -dev > logs/consul.log &
	while true; do sleep 0.5 ; consul members && break ; done

seed:
	go build .
	./hashicorp-game-of-life run

clean:
	pkill nomad consul hashicorp-game-of-life   || true
	while true; do sleep 0.5; consul members    || break ; done
	while true; do sleep 0.5; nomad node status || break ; done
	rm -rf ./logs/

.PHONY: services nomad consul seed clean
