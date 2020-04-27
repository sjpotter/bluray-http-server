package main

import (
	"context"
	"fmt"

	"github.com/jochenvg/go-udev"
)

func main() {
	u := udev.Udev{}
	m := u.NewMonitorFromNetlink("udev")
	m.FilterAddMatchSubsystemDevtype("block", "disk")

	// Create a context
	ctx, _ := context.WithCancel(context.Background())

	// Start monitor goroutine and get receive channel
	ch, _ := m.DeviceChan(ctx)

	fmt.Println("Started listening on channel")
	for d := range ch {
		fmt.Printf("Event: syspath = %v, action = %v, devnode = %v, devpath = %v\n", d.Syspath(), d.Action(), d.Devnode(), d.Devpath())
	}
}
