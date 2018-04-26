# examples

simple redo work

```
package main

import (
	"fmt"
	"github.com/qjpcpu/common/redo"
	"time"
)

func main() {
	job := func() {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "gogogo")
	}
	rep := redo.PerformSafe(redo.WrapFunc(job), time.Second*1)
	go func() {
		time.Sleep(3 * time.Second)
		rep.Stop()
	}()
	rep.Wait()
	fmt.Println("finished")
}
```


more complex, control delay before next loop


```
package main

import (
	"fmt"
	"github.com/qjpcpu/common/redo"
	"time"
)

func main() {
	i := 0
	job := func(ctx *redo.RedoCtx) {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "gogogo")
		i += 1
		if i > 10 {
			ctx.StartNextRightNow() // no delay
			// change delay from default 3 sec to 12 sec
			// ctx.SetDelayBeforeNext(12 * time.Second)
		}
	}
	rep := redo.Perform(job, time.Second*3)
	go func() {
		time.Sleep(3 * time.Second)
		rep.Stop()
	}()
	rep.Wait()
	fmt.Println("finished")
}
```
