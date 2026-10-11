package main

import (
	"context"
	"fmt"
	"goredis/client"
	"log"
	"sync"
	"testing"
	"time"
)


func TestServerWithMultiClient(t *testing.T) {
	
	
	server := NewServer(Config{})

	go func(){
		log.Fatal(server.Start())

	}()
	time.Sleep(time.Second)

	nClients := 10
	wg := sync.WaitGroup{}
	wg.Add(nClients)
	for i:=0; i< nClients;i++{
		go func(i int){
			c, err := client.New("localhost:5001")
			if err != nil {
				log.Fatal(err)
			}
			defer c.Close()

			key := fmt.Sprintf("client_foo_%d", i)
			value := fmt.Sprintf("client_bar_%d", i)

			if err := c.Set(context.TODO(), key, value); err != nil {
				log.Fatal(err)
			}

			val, err := c.Get(context.TODO(), key)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("client %d got this val back =>%s\n",i,  val)
			wg.Done()
		}(i)
	}	

	wg.Wait()
	if len(server.peers) != 0{
		t.Fatalf("expected 0 peers but got %d", len(server.peers))
	}

}