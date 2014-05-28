//  system.go -- nonportable interface functions.
//
//  This file holds system interface functions not usable in the App Engine.

package rsys

import (
	"fmt"
	"rx"
	"syscall"
	"time"
)

//  CPUtime returns the current CPU usage (user time + system time).
func CPUtime() time.Duration {
	var ustruct syscall.Rusage
	rx.CkErr(syscall.Getrusage(0, &ustruct))
	user := time.Duration(syscall.TimevalToNsec(ustruct.Utime))
	sys := time.Duration(syscall.TimevalToNsec(ustruct.Stime))
	return user + sys
	return 0
}

//  Interval returns the CPU time (user + system) since the preceding call.
func Interval() time.Duration {
	total := CPUtime()
	delta := total - prevTotal
	prevTotal = total
	return delta
}

var prevTotal time.Duration // total time at list check

//  ShowInterval calcs and (unless label is empty) prints the last interval.
func ShowInterval(label string) {
	dt := Interval().Seconds()
	if label != "" {
		fmt.Printf("%7.3f %s\n", dt, label)
	}
}
