package main

import (
	"time"
	"log"
)

// Arithmetic over maps of integers
// I feel a bout of APL coming on

func diff(older, newer map[string]uint64) map[string]uint64 {
	accum := make(map[string]uint64)
	
	for k, o := range older {
		// does this key exist in both?
		n, ok := newer[k]
		if !ok {
			continue
		}
		
		// is the newer one less than the older one?  If so, we've wrapped, ignore
		if n < o {
			continue
		}
		
		accum[k] = n - o
	}
	
	return accum
}

func div(data map[string]uint64, by float64) map[string]uint64 {
	accum := make(map[string]uint64)
	for k, v := range data {
		accum[k] = uint64(float64(v) / by)
	}
	return accum
}

func multiply(data map[string]uint64, by uint64) map[string]uint64 {
	accum := make(map[string]uint64)
	for k, v := range data {
		accum[k] = v * by
	}
	return accum
}


func rate(older map[string]uint64, then time.Time, newer map[string]uint64, now time.Time) map[string]uint64 {
	duration := now.Sub(then)
	var divisor float64 = float64(duration) / float64(time.Second)
	log.Print("dividing by %v seconds", divisor)
	
	diffs := diff(older, newer)
	return div(diffs, divisor)
}

