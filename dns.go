package main

import (
	"context"
	"fmt"
	"net"
)

const consulDnsAddr = "127.0.0.1:8600"
const datacenter = "dc1"

func NewDNS() *DNSer {
	return &DNSer{
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

type DNSer struct {
	r   *net.Resolver
	ctx context.Context
}

func (d *DNSer) GetServiceAddr(serviceName string) (addr string, err error) {
	name := fmt.Sprintf("%s.service.%s.consul", serviceName, datacenter)

	_, addrs, err := d.r.LookupSRV(d.ctx, "", "", name)
	if err != nil {
		return "", err
	}
	if len(addrs) < 1 {
		return "", fmt.Errorf("%s: got no addrs :(", name)
	}

	// there should be only one addr in our case.
	target := addrs[0].Target
	port := addrs[0].Port

	ips, err := d.r.LookupHost(d.ctx, target)
	if err != nil {
		return "", err
	}
	if len(ips) < 1 {
		return "", fmt.Errorf("%s: got no ips :(", target)
	}

	return fmt.Sprintf("%s:%d", ips[0], port), nil
}
