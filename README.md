# go GDM package

WiP for the moment

## Usage

```go
import "github.com/cnf/go-gdm"

gdms, err := GetServers()
```

## Available functions
```go
func GetPlayers() ([]*GDMMessage, error)
func GetPlayer(name string) (*GDMMessage, error)
func GetServers() ([]*GDMMessage, error)
func GetServer(name string) (*GDMMessage, error)

func WatchPlayers(freq int) (chan *GDMWatcher, error)
func WatchPlayer(name string) (chan *GDMWatcher, error)
func WatchServers(freq int) (chan *GDMWatcher, error)
func WatchServer(name string) (chan *GDMWatcher, error)
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
w, cerr := WatchServers(5)
if cerr != nil {
    // Error handeling
} else {
    i := 0
    fmt.Printf("%# v\n", w)
    fmt.Println("================")
    for gdm := range w.Watch {
        fmt.Printf("%02d ++++++++++++++\n", i)
        fmt.Printf("%# v\n", gdm)
        i++
        if i >= 10 {
            w.Close()
        }
    }
}
```
