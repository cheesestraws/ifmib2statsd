package main

import (
	"log"
	"time"
	"fmt"

	"github.com/cactus/go-statsd-client/statsd" 
)

var statsdServers []string = []string{"192.168.0.53:8125"}
var targets []string = []string{"192.168.0.1"}
var communities []string = []string{"public"}
var prefix = "network-devices"
var interval time.Duration = 10 * time.Second

// initialise statters from a list of statters
func initStatters(addrs []string, prefix string) []statsd.SubStatter {
	var xs []statsd.SubStatter
	for _, addr := range addrs {
		conf := &statsd.ClientConfig{
			Address: addr,
			Prefix: prefix,
			UseBuffered: true,
			FlushInterval: 300 * time.Millisecond,
		}
		
		cli, err := statsd.NewClientWithConfig(conf)
		if err != nil {
			log.Fatal(err)
		}
		
		xs = append(xs, cli.NewSubStatter(""))
	}
	
	return xs
}

func gauge(ss []statsd.SubStatter, name string, value int64) {
	for _, s := range ss {
		s.Gauge(name, value, 1.0)
	}
}

func main() {
	statters := initStatters(statsdServers, prefix)

	t := time.NewTicker(interval)
	for _ = range t.C {
		for i := range targets {
			ip := targets[i]
			community := communities[i]
			
			results := Poll(ip, community)
			
			// memory usage first
			gauge(statters, fmt.Sprintf("%v.memory.used", results.hostname), int64(results.ramUsed))
			gauge(statters, fmt.Sprintf("%v.memory.avail", results.hostname), int64(results.ramAvail))
			
			// and cpu
			for id, usage := range results.cpuUsage {
				gauge(statters, fmt.Sprintf("%v.cpu.%v", results.hostname, id), int64(usage))
			}
		}
	}
}