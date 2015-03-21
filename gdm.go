// Package gdm provides an interface to the plex.tv GDM protocol for
// Plex player/server discovery.
package gdm

import "net"
import "fmt"
import "strings"
import "time"
import "strconv"

const gdmPlayerPort = 32412
const gdmServerPort = 32414
const serverWaitTime time.Duration = 2

// GetPlayers returns a list of Players.
func GetPlayers() ([]*GDMMessage, error) {
	return getter(gdmPlayerPort)
}

// GetPlayer returns a single Player matching the name supplied.
func GetPlayer(name string) (*GDMMessage, error) {
	gdms, err := getter(gdmPlayerPort)
	if err != nil {
		return nil, err
	}
	for _, gdm := range gdms {
		if (gdm.Props["Name"] == name) || (gdm.Address.IP.String() == name) {
			return gdm, nil
		}
	}
	return nil, fmt.Errorf("no player found named `%s`", name)
}

// GetServers returns a list of Servers.
func GetServers() ([]*GDMMessage, error) {
	return getter(gdmServerPort)
}

// GetServer returns a single Server matching the name supplied.
func GetServer(name string) (*GDMMessage, error) {
	gdms, err := getter(gdmServerPort)
	if err != nil {
		return nil, err
	}
	for _, gdm := range gdms {
		if (gdm.Props["Name"] == name) || (gdm.Address.IP.String() == name) {
			return gdm, nil
		}
	}
	return nil, fmt.Errorf("no player found named `%s`", name)
}

// WatchPlayers returns a *GDMWatcher instance containing a channel
// on which it pushes all Players that answer the regular broadcasts
func WatchPlayers(freq int) (*GDMWatcher, error) {
	gdms, err := watcher(gdmPlayerPort, freq)
	if err != nil {
		return nil, err
	}
	return gdms, nil
}

// WatchServers returns a *GDMWatcher instance containing a channel
// on which it pushes all Servers that answer the regular broadcasts
func WatchServers(freq int) (*GDMWatcher, error) {
	gdms, err := watcher(gdmServerPort, freq)
	if err != nil {
		return nil, err
	}
	return gdms, nil
}

// GDMMessage holds the information of one Player or Server
type GDMMessage struct {
	Address *net.UDPAddr
	Added   bool
	Props   map[string]string
}

// GDMWatcher is the structure containing the Watch channel.
type GDMWatcher struct {
	Watch  chan *GDMMessage
	closer chan bool
}

type gdmBrowser struct {
	mc     chan *GDMMessage
	ticker *time.Ticker
	conn   *net.UDPConn
}

// Close sends all go routines associated a signal to exit.
func (w *GDMWatcher) Close() {
	w.closer <- true
}

func getter(port int) ([]*GDMMessage, error) {
	items := make([]*GDMMessage, 0)
	browser, err := setupBrowser(10)
	if err != nil {
		return nil, nil
	}
	browser.listen()
	browser.browse(port)

	timer := time.NewTimer(time.Second * serverWaitTime)
	done := false
	for !done {
		select {
		case <-timer.C:
			done = true
			browser.conn.Close()
		case msg := <-browser.mc:
			items = append(items, msg)
		}
	}
	return items, nil
}

func watcher(port int, freq int) (*GDMWatcher, error) {
	browser, err := setupBrowser(freq)
	if err != nil {
		return nil, err
	}
	browser.listen()
	browser.browse(port)
	go func() {
		for {
			<-browser.ticker.C
			browser.browse(port)
		}
	}()

	watcher := &GDMWatcher{Watch: browser.mc, closer: make(chan bool)}
	browser.closer(watcher.closer)

	return watcher, nil
}

func setupBrowser(i int) (*gdmBrowser, error) {
	refresh := (time.Duration(i) * time.Second)
	// setup listener to broadcast stuff
	udpaddr, err := net.ResolveUDPAddr("udp4", ":0")
	if err != nil {
		return nil, err
	}
	conn, lerr := net.ListenUDP("udp4", udpaddr)
	if lerr != nil {
		return nil, lerr
	}
	return &gdmBrowser{conn: conn,
		mc:     make(chan *GDMMessage),
		ticker: time.NewTicker(refresh)}, nil
}

func (b *gdmBrowser) listen() {
	go func() {
		buf := make([]byte, 1024)
		for {
			mlen, raddr, err := b.conn.ReadFromUDP(buf)
			if err != nil {
				if nerr, ok := err.(net.Error); !ok || !nerr.Temporary() {
					// this connection has become useless
					close(b.mc)
					return
				}
				continue
			}
			msg := string(buf[0:mlen])
			if strings.HasPrefix(msg, "HTTP/1.0 200 OK") {
				gdmmsg := newGDMMessage(msg, raddr)
				b.mc <- gdmmsg
			}
		}
	}()
}

func (b *gdmBrowser) closer(s chan bool) {
	go func() {
		select {
		case <-s:
			b.conn.Close()
			return
		}
	}()
}

func (b *gdmBrowser) browse(port int) {
	addrs := getBroadcastAddrs()
	ports := strconv.Itoa(port)
	for _, a := range addrs {
		udpdest, err := net.ResolveUDPAddr("udp4", a+":"+ports)
		if err != nil {
			continue
		}
		buf := []byte("M-SEARCH * HTTP/1.1\r\n\r\n")
		_, werr := b.conn.WriteToUDP(buf, udpdest)
		if werr != nil {
			continue
		}
	}
}

func newGDMMessage(data string, addr *net.UDPAddr) *GDMMessage {
	msglines := strings.Split(data, "\r\n")
	gdm := GDMMessage{Address: addr, Added: true, Props: make(map[string]string)}
	for _, m := range msglines {
		if strings.Contains(m, ":") {
			kv := strings.Split(m, ":")
			if len(kv) == 2 {
				gdm.Props[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
			}
		}
	}
	return &gdm
}

func getBroadcastAddrs() (baddr []string) {
	ifaces, _ := net.Interfaces()
	for _, iface := range ifaces {
		// only send on interfaces that are up and can do broadcast
		upandbcast := net.FlagBroadcast | net.FlagUp
		if iface.Flags&upandbcast != upandbcast {
			continue
		}
		addrs, _ := iface.Addrs()
		if len(addrs) == 0 {
			continue
		}
		for _, addr := range addrs {
			var bcast []string
			ip, ipn, err := net.ParseCIDR(addr.String())
			if err != nil {
				continue
			}
			// we only handle ipv4 addresses for now
			ipv4 := ip.To4()
			if ipv4 == nil {
				continue
			}
			for i, m := range ipn.Mask {
				if m == 0 {
					bcast = append(bcast, "255")
				} else {
					bcast = append(bcast, fmt.Sprintf("%d", ipv4[i]))
				}
			}
			baddr = append(baddr, strings.Join(bcast, "."))
		}
	}
	return baddr
}
