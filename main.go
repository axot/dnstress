package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	math_rand "math/rand"
	"net"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
)

var names []string

func main() {
	ctx := context.Background()

	math_rand.Seed(time.Now().UnixNano())

	var namesPath string
	flag.StringVar(&namesPath, "f", namesPath, "path to file containing list of DNS names to resolve")

	flag.Set("logtostderr", "true")
	flag.Parse()

	namesString := `
google.com
yahoo.com
`

	if namesPath != "" {
		b, err := ioutil.ReadFile(namesPath)
		if err != nil {
			glog.Fatalf("error reading %s: %v", namesPath, err)
		}

		namesString = string(b)
	}

	for _, s := range strings.Split(namesString, "\n") {
		s = strings.TrimSpace(s)
		if s != "" {
			names = append(names, s)
		}
	}

	fmt.Printf("dns stress test\n")

	cpus := runtime.NumCPU()

	var wg sync.WaitGroup
	wg.Add(cpus)
	for i := 0; i <= cpus; i++ {
		childCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		go run(childCtx, &wg)
	}

	wg.Wait()
}

func pickHostname() string {
	return names[math_rand.Intn(len(names))]
}

func run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		host := pickHostname()
		err := resolve(ctx, host)

		if err != nil {
			glog.Infof("error resolving host: %v", err)
			continue
		}

		// At most 100 qps (per pod)
		time.Sleep(10 * time.Millisecond)
	}
}

func resolve(ctx context.Context, host string) error {
	start := time.Now()
	ips, err := lookupIP(ctx, host)
	elapsed := time.Since(start)

	if err != nil {
		return fmt.Errorf("error from lookup: %v, time elapsed: %s", err, elapsed)
	}

	if len(ips) == 0 {
		return fmt.Errorf("no ips from lookup of %s", host)
	}

	if elapsed > time.Second*1 {
		return fmt.Errorf("take to long time from lookup: %s, time elapsed: %s", host, elapsed)
	}
	return nil
}

// LookupIP looks up host using the local resolver.
// It returns a slice of that host's IPv4 and IPv6 addresses.
func lookupIP(ctx context.Context, host string) ([]net.IP, error) {
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	ips := make([]net.IP, len(addrs))
	for i, ia := range addrs {
		ips[i] = ia.IP
	}
	return ips, nil
}
