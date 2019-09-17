package main

import (
	"fmt"
	"time"
	"strings"

	"github.com/soniah/gosnmp"
)

func ifTable(column int) string {
	return fmt.Sprintf(".1.3.6.1.2.1.2.2.1.%d", column)
}

func ifxTable(column int) string {
	return fmt.Sprintf("1.3.6.1.2.1.31.1.1.1.%d", column)
}

func oididx(oid string) string {
	s := strings.Split(oid, ".")
	return s[len(s)-1]
}

type PollResults struct {
	when time.Time
	hostname string
	statname string
	
	cpuUsage map[string]int
	ramTotal uint64
	ramUsed uint64
	ramAvail uint64
	
	ifName map[string]string
	ifType map[string]string
	ifHCInOctets map[string]uint64
	ifHCOutOctets map[string]uint64
	ifInErrors map[string]uint64
	ifOutErrors map[string]uint64
}

func Poll(ip string, community string) *PollResults {
	gosnmp.Default.Target = ip
	gosnmp.Default.Community = community
	gosnmp.Default.Timeout = time.Duration(5 * time.Second)
	gosnmp.Default.Version = gosnmp.Version2c
	
	err := gosnmp.Default.Connect()
	if err != nil {
		fmt.Printf("cannot connect to %v: %v\n", ip,err)
		return nil
	}
	defer gosnmp.Default.Conn.Close()
	
	fmt.Printf("foo\n")
	
	// hostname
	var hostname string
	pdus, err := gosnmp.Default.BulkWalkAll("1.3.6.1.2.1.1.5")
	if err != nil {
		fmt.Printf("cannot query %v: %v\n", ip,err)
		return nil
	} else {
		for _, p := range pdus {
			hostname = string(p.Value.([]byte))
		}
	}
	
	// ram?
	// first find the id in the HOST-RESOURCES-MIB storage table for the RAM
	var ramStorageID string
	pdus, err = gosnmp.Default.BulkWalkAll("1.3.6.1.2.1.25.2.3.1.2")
	if err != nil {
		fmt.Printf("cannot query %v: %v\n", ip,err)
		return nil
	} else {
		for _, p := range pdus {
			if p.Value.(string) == ".1.3.6.1.2.1.25.2.1.2" {
				ramStorageID = oididx(p.Name)
			}
		}
	}
	
	var ramTotal uint64
	var ramUsed uint64
	var ramAvail uint64
	if ramStorageID != "" {
		// find the allocation unit size
		var allocationUnitSize uint64
		pdus, err = gosnmp.Default.BulkWalkAll("1.3.6.1.2.1.25.2.3.1.4")
		if err != nil {
			fmt.Printf("cannot query %v: %v\n", ip,err)
			return nil
		} else {
			for _, p := range pdus {
				if oididx(p.Name) == ramStorageID {
					allocationUnitSize = uint64(p.Value.(int))
					fmt.Printf("Allocation unit size: %v\n", p.Value.(int))
				}
			}
		}

		
		// now find the total
		pdus, err = gosnmp.Default.BulkWalkAll("1.3.6.1.2.1.25.2.3.1.5")
		if err != nil {
			fmt.Printf("cannot query %v: %v\n", ip,err)
			return nil
		} else {
			for _, p := range pdus {
				if oididx(p.Name) == ramStorageID {
					ramTotal = uint64(p.Value.(int)) * allocationUnitSize
					fmt.Printf("Ram total: %v\n", ramTotal)
				}
			}
		}
		
		// and the used
		pdus, err = gosnmp.Default.BulkWalkAll("1.3.6.1.2.1.25.2.3.1.6")
		if err != nil {
			fmt.Printf("cannot query %v: %v\n", ip,err)
			return nil
		} else {
			for _, p := range pdus {
				if oididx(p.Name) == ramStorageID {
					ramUsed = uint64(p.Value.(int)) * allocationUnitSize
					fmt.Printf("Ram used: %v\n", ramUsed)
				}
			}
		}

		ramAvail = ramTotal-ramUsed
	}
	
	// cpu
	cpuUsage := make(map[string]int)
	pdus, err = gosnmp.Default.BulkWalkAll("1.3.6.1.2.1.25.3.3.1.2")
	if err != nil {
		fmt.Printf("cannot query %v: %v\n", ip,err)
		return nil
	} else {
		for _, p := range pdus {
			cpuUsage[oididx(p.Name)] = p.Value.(int)
			fmt.Printf("cpu %v => %d\n", oididx(p.Name), cpuUsage[oididx(p.Name)])
		}
	}

	
	// names
	ifName := make(map[string]string)
	pdus, err = gosnmp.Default.BulkWalkAll(ifxTable(1))
	if err != nil {
		fmt.Printf("cannot query %v: %v\n", ip,err)
		return nil
	} else {
		fmt.Printf("bar %v\n", ifTable(2))
		for _, p := range pdus {
			ifName[oididx(p.Name)] = string(p.Value.([]byte))
			fmt.Printf("%v => %v\n", oididx(p.Name), string(p.Value.([]byte)))
		}
	}
	
	// types
	ifType := make(map[string]string)
	pdus, err = gosnmp.Default.BulkWalkAll(ifTable(3))
	if err != nil {
		fmt.Printf("cannot query %v: %v\n", ip,err)
		return nil
	} else {
		for _, p := range pdus {
			ifType[ifName[oididx(p.Name)]] = ianaIFType[p.Value.(int)]
			fmt.Printf("%v => %v\n", oididx(p.Name), ianaIFType[p.Value.(int)])
		}
	}
	
	// ifHCInOctets
	ifHCInOctets := make(map[string]uint64)
	pdus, err = gosnmp.Default.BulkWalkAll(ifxTable(6))
	if err != nil {
		fmt.Printf("cannot query %v: %v\n", ip,err)
		return nil
	} else {
		for _, p := range pdus {
			ifHCInOctets[ifName[oididx(p.Name)]] = p.Value.(uint64)
			fmt.Printf("%v => %v\n", ifName[oididx(p.Name)], p.Value.(uint64))
		}
	}
	
	// ifHCOutOctets
	ifHCOutOctets := make(map[string]uint64)
	pdus, err = gosnmp.Default.BulkWalkAll(ifxTable(10))
	if err != nil {
		fmt.Printf("cannot query %v: %v\n", ip,err)
		return nil
	} else {
		for _, p := range pdus {
			ifHCOutOctets[ifName[oididx(p.Name)]] = p.Value.(uint64)
			fmt.Printf("%v => %v\n", oididx(p.Name), p.Value.(uint64))
		}
	}

	// ifInErrors
	ifInErrors := make(map[string]uint64)
	pdus, err = gosnmp.Default.BulkWalkAll(ifTable(14))
	if err != nil {
		fmt.Printf("cannot query %v: %v\n", ip,err)
		return nil
	} else {
		for _, p := range pdus {
			ifInErrors[ifName[oididx(p.Name)]] = uint64(p.Value.(uint))
			fmt.Printf("%v => %v\n", oididx(p.Name), p.Value.(uint))
		}
	}
	
	// ifOutErrors
	ifOutErrors := make(map[string]uint64)
	pdus, err = gosnmp.Default.BulkWalkAll(ifTable(20))
	if err != nil {
		fmt.Printf("cannot query %v: %v\n", ip,err)
		return nil
	} else {
		for _, p := range pdus {
			ifOutErrors[ifName[oididx(p.Name)]] = uint64(p.Value.(uint))
			fmt.Printf("%v => %v\n", oididx(p.Name), p.Value.(uint))
		}
	}
	
	
	return &PollResults{
		when: time.Now(),
		hostname: hostname,
		statname: strings.Replace(hostname, ".", "_", -1),
		
		cpuUsage: cpuUsage,
		ramTotal: ramTotal,
		ramUsed: ramUsed,
		ramAvail: ramAvail,
		
		ifName: ifName,
		ifType: ifType,
		ifHCInOctets: ifHCInOctets,
		ifHCOutOctets: ifHCOutOctets,
		ifInErrors: ifInErrors,
		ifOutErrors: ifOutErrors,
	}
}