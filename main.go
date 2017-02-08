package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"os"

	"log"

	"github.com/coreos/etcd/clientv3"
)

var (
	timeout = 5 * time.Second
)

func main() {

	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379", "localhost:22379", "localhost:32379"},
		DialTimeout: timeout,
	})
	if err != nil {
		// handle error!
	}
	defer cli.Close()

	rch := cli.Watch(context.Background(), "/key", clientv3.WithPrefix())
	go func() {
		for wresp := range rch {
			for _, ev := range wresp.Events {
				fmt.Printf("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
			}
		}
	}()

	for i := range make([]int, 3) {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		_, err = cli.Put(ctx, fmt.Sprintf("/key_%d", i), "value")
		cancel()
		if err != nil {
			log.Fatal(err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	resp, err := cli.Get(ctx, "/key", clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
	cancel()
	if err != nil {
		log.Fatal(err)
	}
	for _, ev := range resp.Kvs {
		fmt.Printf("%s : %s\n", ev.Key, ev.Value)
	}

	<-sigs
}
