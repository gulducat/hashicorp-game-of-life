package main

import (
	"context"
	"fmt"
	"net"
	"strings"
)

const datacenter = "dc1"

func NewConsulDNS() *ConsulDNS {
	consulDnsAddr := strings.Replace(ConsulAddr, "8500", "8600", 1)
	consulDnsAddr = strings.Replace(consulDnsAddr, "http://", "", 1)
	consulDnsAddr = strings.Replace(consulDnsAddr, "https://", "", 1)
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
	name := serviceName + ".service." + datacenter + ".consul"

	_, addrs, err := c.r.LookupSRV(c.ctx, "", "", name)
	if err != nil {
		return "", fmt.Errorf("Error getting SRV record for %s: %w", name, err)
	}
	if len(addrs) < 1 {
		return "", fmt.Errorf("%s: got no addrs :(", name)
	}

	// there should be only one addr in our case.
	target := addrs[0].Target
	port := addrs[0].Port

	ips, err := c.r.LookupHost(c.ctx, target)
	if err != nil {
		return "", fmt.Errorf("Error looking up host for %s: %w", target, err)
	}
	if len(ips) < 1 {
		return "", fmt.Errorf("%s: got no ips :(", target)
	}

	return fmt.Sprintf("%s:%d", ips[0], port), nil
}
