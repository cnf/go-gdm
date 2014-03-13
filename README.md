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
```
And not implemented yet:
```go
func WatchPlayers(freq int) (chan *GDMMessage, error)
func WatchPlayer(name string) (chan *GDMMessage, error)
func WatchServers(freq int) (chan *GDMMessage, error)
func WatchServer(name string) (chan *GDMMessage, error)
```

### GDMMessage
```go
type GDMMessage struct {
    Address *net.UDPAddr
    Added bool
    Props map[string]string
}
```
