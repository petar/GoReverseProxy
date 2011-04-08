
package main

import (
	"container/list"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
	"github.com/petar/GoHTTP/http"
	"github.com/petar/GoHTTP/server"
	"github.com/petar/GoHTTP/util"
)

type Proxy struct {
	sync.Mutex // protects listen and conns

	listen net.Listener
	fdl    util.FDLimiter
	pairs  map[*connPair]int
	ech    chan os.Error

	config	*Config
}

type connPair struct {
	s	*server.StampedServerConn
	c	*server.StampedClientConn
}

func (cp *connPair) GetStamp() int64 { return min64(cp.s.GetStamp(), cp.c.GetStamp()) }

func min64(p,q int64) int64 {
	if p < q {
		return p
	}
	return q
}

func NewProxy(l net.Listener, config *Config) (*Proxy, os.Error) {
	p := &Proxy{
		listen: l,
		config: config,
		pairs:  make(map[*connPair]int),
		ech:    make(chan os.Error),
	}
	p.fdl.Init(config.FDLimit)
	return p, nil
}

func NewProxyEasy(addr, configfile string) (*Proxy, os.Error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	conf, err := ParseConfigFile(configfile)
	if err != nil {
		return nil, err
	}
	return NewProxy(l, conf)
}

func (p *Proxy) Start() os.Error {
	go p.acceptLoop()
	go p.expireLoop()
	return <-p.ech
}

func (p *Proxy) ConfigString() string { return p.config.String() }

func (p *Proxy) expireLoop() {
	for i := 0; ; i++ {
		p.Lock()
		if p.listen == nil {
			p.Unlock()
			return
		}
		now := time.Nanoseconds()
		kills := list.New()
		for q, _ := range p.pairs {
			if now - q.GetStamp() >= p.config.Timeout {
				kills.PushBack(q)
			}
		}
		p.Unlock()
		elm := kills.Front()
		for elm != nil {
			q := elm.Value.(*connPair)
			p.bury(q)
			elm = elm.Next()
		}
		kills.Init()
		kills = nil
		time.Sleep(p.config.Timeout)
	}
}

func (p *Proxy) acceptLoop() {
	for {
		p.Lock()
		l := p.listen
		p.Unlock()
		if l == nil {
			return
		}
		p.fdl.Lock()
		c, err := l.Accept()
		if err != nil {
			log.Printf("Error accepting: %s\n", err)
			if c != nil {
				c.Close()
			}
			p.fdl.Unlock()
			p.ech <- err
			return
		}
		go p.connLoop(c)
	}
}

// prepConn() takes a net.Conn and attaches a file descriptor release in its Close method
func (p *Proxy) prepConn(c net.Conn) (net.Conn, os.Error) {
	c.(*net.TCPConn).SetKeepAlive(true)
	err := c.SetReadTimeout(p.config.Timeout)
	if err != nil {
		log.Printf("Error TCP set read timeout: %s\n", err)
		c.Close()
		p.fdl.Unlock()
		return c, err
	}
	err = c.SetWriteTimeout(p.config.Timeout)
	if err != nil {
		log.Printf("Error TCP set write timeout: %s\n", err)
		c.Close()
		p.fdl.Unlock()
		return c, err
	}
	return util.NewRunOnCloseConn(c, func() { p.fdl.Unlock() }), nil
}

func (p *Proxy) connLoop(s_ net.Conn) {
	s_, err := p.prepConn(s_)
	if err != nil {
		return
	}
	st := server.NewStampedServerConn(s_, nil)

	// Read and parse first request
	req0, err := st.Read()
	if err != nil {
		log.Printf("Read first Request: %s\n", err)
		st.Close()
		return
	}
	req0.Host = strings.ToLower(strings.TrimSpace(req0.Host))
	if req0.Host == "" {
		st.Write(req0, http.NewResponse400String("GoFrontline: missing host"))
		st.Close()
		return
	}

	// Connect to host
	host := p.config.ActualHost(req0.Host)
	if host == "" {
		st.Write(req0, http.NewResponse400String("GoFrontline: unknwon host"))
		st.Close()
		return
	}
	p.fdl.Lock()
	c_, err := net.Dial("tcp", host)
	if err != nil {
		log.Printf("Dial server: %s\n", err)
		if c_ != nil {
			c_.Close()
		}
		p.fdl.Unlock()
		st.Write(req0, http.NewResponse400String("GoFrontline: error dialing host"))
		st.Close()
		return
	}
	c_, err = p.prepConn(c_)
	if err != nil {
		st.Write(req0, http.NewResponse400String("GoFrontline: error on host conn"))
		st.Close()
		return
	}
	ct := server.NewStampedClientConn(c_, nil)
	q := p.register(st, ct)

	ch := make(chan *http.Request, 5)
	go p.backLoop(ch, q)
	p.frontLoop(ch, q, req0)
}

// Read request from browser, write request to server, notify backLoop and repeat
func (p *Proxy) frontLoop(ch chan<- *http.Request, q *connPair, req0 *http.Request) {
	var req *http.Request = req0
	for {
		// Read request from browser
		if req == nil {
			var err os.Error
			req, err = q.s.Read()
			if err != nil {
				log.Printf("Read Request: %s\n", err)
				goto __Close
			}
			// TODO: Verify same Host
		}
		shouldClose := req.Close

		err := q.c.Write(req)
		if err != nil {
			log.Printf("Write Request: %s\n", err)
			goto __Close
		}
		ch <- req

		if shouldClose {
			goto __Close
		}
		req = nil
	}
__Close:
	close(ch)
}

// Read request from frontLoop, read response from server, send response to browser, repeat
func (p *Proxy) backLoop(ch <-chan *http.Request, q *connPair) {
	for {
		req, closed := <-ch
		if closed && req == nil {
			goto __Close
		}

		resp, err := q.c.Read(req)
		if err != nil {
			log.Printf("Read Response: %s\n", err)
			goto __Close
		}

		err = q.s.Write(req, resp)
		if err != nil {
			log.Printf("Write Response: %s\n", err)
			goto __Close
		}
	}
__Close:
	p.bury(q)
}

func (p *Proxy) register(st *server.StampedServerConn, ct *server.StampedClientConn) *connPair {
	p.Lock()
	defer p.Unlock()

	q := &connPair{st,ct}
	p.pairs[q] = 1
	return q
}

func (p *Proxy) bury(q *connPair) {
	p.Lock()
	defer p.Unlock()

	p.pairs[q] = 0, false
	q.s.Close()
	q.c.Close()
}

