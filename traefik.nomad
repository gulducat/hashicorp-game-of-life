locals {
  http = 80
}

variable "consul_http_addr" {}

job "traefik" {
  datacenters = ["dc1"]
  type        = "system"

  group "traefik" {
    network {
      port "web" {
        static = local.http
      }
    }
    task "traefik" {
      driver         = "docker"
      config {
        image        = "traefik:v2.2"
        ports = ["web"]
        volumes = [
          "local/traefik.yml:/etc/traefik/traefik.yml",
        ]
      }

      env {
        CONSUL_HTTP_ADDR = var.consul_http_addr
      }

      resources {
        cpu    = 100
        memory = 128
      }

      service {
        name         = "traefik"
        port         = "web"
        tags = [
          "traefik.enable=true",
          "traefik.http.routers.traefik.tls=false",
          "traefik.http.routers.traefik.service=api@internal",
          "traefik.http.routers.traefik.rule=PathPrefix(`/api`, `/dashboard`)",
        ]
      }

      template {
        destination = "local/traefik.yml"
        data = <<CONF_YAML
accessLog: false

api:
  dashboard: true

ping:
  entryPoint: "web"

entryPoints:
  web:
    address: ":{{ env "NOMAD_PORT_web" }}"

log:
  format: json
  level: debug

serversTransport:
  insecureSkipVerify: true

providers:
  consulCatalog:
    endpoint:
      address: {{ env "CONSUL_HTTP_ADDR" }}
      scheme: http
    exposedByDefault: false
    prefix: traefik

CONF_YAML
      }
    }
  }
}
