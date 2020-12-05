package goSam

import (
	"context"
	"log"
	"net"
	"strings"
)

// DialContext implements the net.DialContext function and can be used for http.Transport
func (c *Client) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	c.oml.Lock()
	defer c.oml.Unlock()
	errCh := make(chan error, 1)
	connCh := make(chan net.Conn, 1)
	go func() {
		if conn, err := c.DialContextFree(network, addr); err != nil {
			errCh <- err
		} else if ctx.Err() != nil {
			log.Println(ctx)
			errCh <- ctx.Err()
		} else {
			connCh <- conn
		}
	}()
	select {
	case err := <-errCh:
		return nil, err
	case conn := <-connCh:
		return conn, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *Client) dialCheck(addr string) (int32, bool) {
	if c.lastaddr == "invalid" {
		return c.NewID(), true
		//	} else if addr == "" {
		//		return c.id, false
	} else if c.lastaddr != addr {
		return c.NewID(), true
	}
	return c.NewID(), false
}

func (c *Client) Dial(network, addr string) (net.Conn, error) {
	return c.DialContextFree(network, addr)
}

// Dial implements the net.Dial function and can be used for http.Transport
func (c *Client) DialContextFree(network, addr string) (net.Conn, error) {
	c.ml.Lock()
	defer c.ml.Unlock()
	portIdx := strings.Index(addr, ":")
	if portIdx >= 0 {
		addr = addr[:portIdx]
	}
	addr, err := c.Lookup(addr)
	if err != nil {
		log.Printf("LOOKUP DIALER ERROR %s %s", addr, err)
		return nil, err
	}

	// var test bool
	//c.id, _ = c.dialCheck(addr)
	// log.Println("Address indicates we need to dial a new session.")
	//	if test {
	c.destination, err = c.CreateStreamSession(c.id, c.destination)
	if err != nil {
		c, err = c.NewClient(c.id + 1)
		if err != nil {
			return nil, err
		}
		c.destination, err = c.CreateStreamSession(c.id, c.destination)
		if err != nil {
			return nil, err
		}
	}
	//	}
	c, err = c.NewClient(c.id)
	if err != nil {
		return nil, err
	}
	c.lastaddr = addr
	err = c.StreamConnect(c.id, addr)
	if err != nil {
		return nil, err
	}
	//}
	return c.SamConn, nil
}
