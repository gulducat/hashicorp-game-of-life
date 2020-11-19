package main

import (
	"context"
	"fmt"
	"net"
)

const consulDnsAddr = "127.0.0.1:8600"
const datacenter = "dc1"

func NewConsulDNS() *ConsulDNS {
	return &ConsulDNS{
		r: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{}
				return d.DialContext(ctx, "udp", consulDnsAddr)
			},
		},
		ctx: context.Background(),
	}
}

type ConsulDNS struct {
	r   *net.Resolver
	ctx context.Context
}

func (c *ConsulDNS) GetServiceAddr(serviceName string) (addr string, err error) {
	name := fmt.Sprintf("%s.service.%s.consul", serviceName, datacenter)

	_, addrs, err := c.r.LookupSRV(c.ctx, "", "", name)
	if err != nil {
		return "", err
	}
	if len(addrs) < 1 {
		return "", fmt.Errorf("%s: got no addrs :(", name)
	}

	// there should be only one addr in our case.
	target := addrs[0].Target
	port := addrs[0].Port

	ips, err := c.r.LookupHost(c.ctx, target)
	if err != nil {
		return "", err
	}
	if len(ips) < 1 {
		return "", fmt.Errorf("%s: got no ips :(", target)
	}

	return fmt.Sprintf("%s:%d", ips[0], port), nil
}
