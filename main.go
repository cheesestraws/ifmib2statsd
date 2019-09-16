package main

import (
	"log"
	"time"
	"fmt"

	"github.com/cactus/go-statsd-client/statsd" 
)

var statsdServers []string = []string{"192.168.0.53:8125"}
var targets []string = []string{"192.168.0.1", "192.168.0.254", "ap.lan", "pf.speakersassociates.com"}
var communities []string = []string{"public", "public", "public", "public"}
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
	oldMeasurements := make(map[string]*PollResults)

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
			
			// Do we have an old measurement for this target?  If so, we can calculate
			// some rates.
			old := oldMeasurements[ip]
			if old == nil {
				oldMeasurements[ip] = results
				continue
			}
			
			// yes!  let's do it
			ifOutBps := multiply(rate(old.ifHCOutOctets, old.when, results.ifHCOutOctets, results.when), 8)
			for k, v := range ifOutBps {
				gauge(statters, fmt.Sprintf("%v.interfaces.%v.%v.out", results.hostname, results.ifType[k], k), int64(v))
			}
			ifInBps := multiply(rate(old.ifHCInOctets, old.when, results.ifHCInOctets, results.when), 8)
			for k, v := range ifInBps {
				gauge(statters, fmt.Sprintf("%v.interfaces.%v.%v.in", results.hostname, results.ifType[k], k), int64(v))
			}
			ifOutErrors := rate(old.ifOutErrors, old.when, results.ifOutErrors, results.when)
			for k, v := range ifOutErrors {
				gauge(statters, fmt.Sprintf("%v.interfaces.%v.%v.out-errors", results.hostname, results.ifType[k], k), int64(v))
			}
			ifInErrors := rate(old.ifOutErrors, old.when, results.ifInErrors, results.when)
			for k, v := range ifInErrors {
				gauge(statters, fmt.Sprintf("%v.interfaces.%v.%v.in-errors", results.hostname, results.ifType[k], k), int64(v))
			}



			
			oldMeasurements[ip] = results

		}
	}
}