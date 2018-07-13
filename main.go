package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/hashicorp/memberlist"
)

type bootNodes []string

func (bn *bootNodes) String() string {
	return strings.Join(*bn, ", ")
}

func (bn *bootNodes) Set(value string) error {
	*bn = append(*bn, value)
	return nil
}

const (
	dirMode   = 0755
	fileMode  = 0644
	fileFlags = os.O_CREATE | os.O_RDWR

	// NumReplicas default number of replicas.
	NumReplicas uint8 = 2 // 32

	// NumShards default number of shards.
	NumShards uint16 = 8 // 1024
)

var (
	nothing struct{}

	host     string
	port     uint
	boots    bootNodes
	basePath string
)

func init() {
	flag.StringVar(&host, "host", "0.0.0.0", "Hostname or IP address")
	flag.UintVar(&port, "port", 50000, "Port number")
	flag.Var(&boots, "boot", "Boot node address(es)")
	flag.StringVar(&basePath, "base", "./data", "Base path")
	flag.Parse()
}

func main() {
	mutex := &sync.Mutex{}

	fs := NewFileSystem()
	name := fmt.Sprintf("%s:%d", host, port)
	ring := NewRing(NumReplicas, name)
	sharder := NewSharder(NumShards, fs, basePath)
	cache := NewCache()
	events := NewEvents(mutex, ring, sharder, cache)
	_ = NewAPI(mutex, ring, sharder, cache, fs)

	cfg := memberlist.DefaultLocalConfig()
	cfg.Name = name
	cfg.BindAddr = host
	cfg.BindPort = int(port)
	cfg.Events = events
	// cfg.ProtocolVersion = memberlist.ProtocolVersionMax

	list, err := memberlist.Create(cfg)
	if err != nil {
		panic(err)
	}

	if len(boots) > 0 {
		_, err = list.Join(boots)
		if err != nil {
			panic(err)
		}
	}

	fmt.Printf("SELF: %s\n", list.LocalNode().Name)

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGTERM)
	<-sig
}
