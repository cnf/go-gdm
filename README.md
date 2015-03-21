# go GDM package

[go](http://golang.org) package to browse [Plex](http://plex.tv) server / players with GDM.

Based on the work, and with the help of Tobias Hieta <tobias@plexapp.com>

## Documentation

[![godoc](http://godoc.org/github.com/cnf/go-gdm?status.png)](http://godoc.org/github.com/cnf/go-gdm)

## gdmbrowser

To install the gdmbrowser binary run, including the `...` at the end

    go get github.com/cnf/go-gdm/...

If you have `$GOPATH/bin` in your path, you can then run `gdmbrowser -h`

## Available functions
```go
func GetPlayers() ([]*GDMMessage, error)
func GetPlayer(name string) (*GDMMessage, error)
func GetServers() ([]*GDMMessage, error)
func GetServer(name string) (*GDMMessage, error)

func WatchPlayers(freq int) (chan *GDMWatcher, error)
func WatchServers(freq int) (chan *GDMWatcher, error)
```

### GDMMessage
```go
type GDMMessage struct {
    Address *net.UDPAddr
    Added bool
    Props map[string]string
}
```

### GDMWatcher
```go
type GDMWatcher struct {
    Watch chan *GDMMessage
    closer chan bool
}

func (w *GDMWatcher) Close()
```

## Examples

```go

box, err := gdm.GetServer("someName")
if err != nil {
    fmt.Println(err.Error())
} else {
    fmt.Printf("%# v\n", box.Props["Name"])
}
```

```go
w, cerr := gdm.WatchServers(5)
if cerr != nil {
    // Error handling
} else {
    i := 0
    fmt.Printf("%# v\n", w)
    fmt.Println("================")
    for gdm := range w.Watch {
        fmt.Printf("%02d ++++++++++++++\n", i)
        fmt.Printf("%# v\n", gdm)
        i++
        if i >= 10 {
            w.Close() // Call close to quit to clean up.
        }
    }
}
```
