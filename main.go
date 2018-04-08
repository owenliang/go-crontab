package main

import (
	"github.com/gorhill/cronexpr"
	"time"
	"fmt"
)

func main()  {
	now := time.Now()
	triggerTime := cronexpr.MustParse("* * * *  *").Next(now)
	fmt.Println(now, triggerTime)
}
