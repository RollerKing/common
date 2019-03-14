package election

import (
	"fmt"
	"github.com/qjpcpu/log"
	"time"
)

// Example for ha
func Example() {
	// close log for better fmt.Println
	log.GetBuilder().SetTypedLevel(log.CRITICAL).Submit()

	finishExampleC := make(chan struct{}, 1)
	var commonEtcdKey = "/share-key"
	endpoints := []string{"127.0.0.1:2379"}
	// server1 on host1
	server1 := New(endpoints, commonEtcdKey)
	fmt.Println("server1 started, wait to be leader...")
	downOne := 0
	go server1.Start()
	go func() {
		for {
			select {
			case role := <-server1.RoleC():
				if role == Leader {
					fmt.Printf("server1 switch to %s, I can work now\n", role.String())
					if downOne == 0 {
						downOne = 1
						time.Sleep(1 * time.Second)
						server1.Stop() // simulate server is down unexpected
						fmt.Println("server1 is down unexpected!")
						return
					} else {
						close(finishExampleC)
					}
				} else {
					fmt.Printf("server1 switch to %s, I cant work\n", role.String())
				}
			case <-finishExampleC:
				return
			}
		}
	}()
	// server2 on another host
	server2 := New(endpoints, commonEtcdKey)
	fmt.Println("server2 started, wait to be leader...")
	go server2.Start()
	go func() {
		for {
			select {
			case role := <-server2.RoleC():
				if role == Leader {
					fmt.Printf("server2 switch to %s, I can work now\n", role.String())
					if downOne == 0 {
						downOne = 1
						time.Sleep(1 * time.Second)
						server2.Stop() // simulate server is down unexpected
						fmt.Println("server2 is down unexpected!")
						return
					} else {
						close(finishExampleC)
					}
				} else {
					fmt.Printf("server2 switch to %s, I cant work\n", role.String())
				}
			case <-finishExampleC:
				return
			}
		}
	}()
	<-finishExampleC
	fmt.Println("example finished.")
}
