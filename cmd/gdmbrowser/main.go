package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/cnf/go-gdm"
)

var server bool
var player bool
var watch bool
var interval int

func main() {
	flag.BoolVar(&server, "s", false, "look for servers")
	flag.BoolVar(&player, "p", false, "look for players")
	flag.BoolVar(&watch, "w", false, "continue watching for new entries")
	flag.IntVar(&interval, "i", 5, "watch polling interval in seconds")

	flag.Parse()

	sigc := make(chan os.Signal, 1)
	stop := make(chan bool)
	signal.Notify(sigc, os.Interrupt)
	go func() {
		<-sigc
		close(stop)
	}()

	if !player && !server {
		server = true
	}

	if player {
		if watch {
			watchPlayers(stop, interval)
		} else {
			getPlayers()
		}
	}

	if server {
		if watch {
			watchServers(stop, interval)
		} else {
			getServers()
		}
	}

	if watch {
		<-stop
	}

}

func getPlayers() {
	fmt.Printf("# players ##########\n")
	players, err := gdm.GetPlayers()
	if err != nil {
		fmt.Printf("%s", err.Error())
		return
	}
	for _, v := range players {
		prettyPlexPrint(v)
	}
}

func getServers() {
	fmt.Printf("# servers ##########\n")
	servers, err := gdm.GetServers()
	if err != nil {
		fmt.Printf("%s", err.Error())
		return
	}
	for _, v := range servers {
		prettyPlexPrint(v)
	}
}

func watcher(stop chan bool, w *gdm.GDMWatcher) {
	go func() {
		for {
			select {
			case msg := <-w.Watch:
				prettyPlexPrint(msg)
			case <-stop:
				w.Close()
				return
			}
		}
	}()
}

func watchPlayers(stop chan bool, i int) {
	wp, err := gdm.WatchPlayers(i)
	if err != nil {
		fmt.Printf("%s", err.Error())
		return
	}
	watcher(stop, wp)
}

func watchServers(stop chan bool, i int) {
	wp, err := gdm.WatchServers(i)
	if err != nil {
		fmt.Printf("%s", err.Error())
		return
	}
	watcher(stop, wp)
}

func prettyPlexPrint(msg *gdm.GDMMessage) {
	fmt.Printf("Name: %s\n", msg.Props["Name"])
	fmt.Printf("  Address: %s:%d\n", msg.Address.IP, msg.Address.Port)
	if msg.Address.Zone != "" {
		fmt.Printf("    Zone: %s", msg.Address.Zone)
	}
	fmt.Printf("  Added: %t\n", msg.Added)
	for k, v := range msg.Props {
		fmt.Printf("      %-25s : %s\n", k, v)
	}
	fmt.Print("\n")
}
