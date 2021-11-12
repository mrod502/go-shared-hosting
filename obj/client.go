package obj

import (
	"errors"
	"net"
	"net/http"
	"time"

	gocache "github.com/mrod502/go-cache"
	"go.uber.org/atomic"
)

var (
	ErrTooManyRequests        = errors.New("too many requests")
	ErrBanned                 = errors.New("fuck off")
	ErrNotFound               = errors.New("not found")
	MaxRatelimits      uint32 = 10
	BanLapseDuration          = time.Hour * 24 * 7
	MaxNumBans         uint32 = 3
)

func NewClient(maxReqs uint32, window time.Duration, remoteAddr net.IPAddr) *Client {
	return &Client{
		RemoteAddr:      remoteAddr,
		LastRequest:     atomic.NewTime(time.Now()),
		RatelimitWindow: window,
		ReqsSinceReset:  atomic.NewUint32(0),
		NumRatelimits:   atomic.NewUint32(0),
		MaxRequests:     maxReqs,
		Banned:          atomic.NewBool(false),
		TimesBanned:     atomic.NewUint32(0),
		PermaBanned:     atomic.NewBool(false),
	}
}

type Client struct {
	RemoteAddr      net.IPAddr
	LastRequest     *atomic.Time
	RatelimitWindow time.Duration
	ReqsSinceReset  *atomic.Uint32
	NumRatelimits   *atomic.Uint32
	MaxRequests     uint32
	Banned          *atomic.Bool
	TimesBanned     *atomic.Uint32
	PermaBanned     *atomic.Bool
}

func (c *Client) resetRequestCount() {
	time.Sleep(c.RatelimitWindow)
	c.ReqsSinceReset.Store(0)
}

func (c *Client) Authorize(r *http.Request) (bool, error) {
	if c.isBanned() {
		return false, ErrBanned
	}
	if c.ReqsSinceReset.Load() == 0 {
		c.resetRequestCount()
	}

	if c.ReqsSinceReset.Load() >= c.MaxRequests {
		if c.NumRatelimits.Inc() > MaxRatelimits {
			c.Ban()
		}
		return false, ErrTooManyRequests
	}

	c.ReqsSinceReset.Inc()
	c.LastRequest.Store(time.Now())

	return true, nil
}

func (c Client) isBanned() bool {
	return c.Banned.Load() || c.PermaBanned.Load()
}

func (c *Client) Ban() {
	c.Banned.Store(true)
	if c.TimesBanned.Inc() > MaxNumBans {
		c.PermaBanned.Store(true)
	}
}

type ClientCache struct {
	v           *gocache.InterfaceCache
	MaxRequests uint32
	RLimWindow  time.Duration
}

func NewClientCache(maxReqs uint32, rlimWindow time.Duration) *ClientCache {
	return &ClientCache{
		v:           gocache.NewInterfaceCache(),
		MaxRequests: maxReqs,
		RLimWindow:  rlimWindow,
	}
}

func (c *ClientCache) Authorize(r *http.Request) (bool, error) {
	cli, err := c.Get(r.RemoteAddr)
	if err != nil {
		ip := net.ParseIP(r.RemoteAddr)
		cli = NewClient(c.MaxRequests, c.RLimWindow, net.IPAddr{IP: ip})
	}

	return cli.Authorize(r)
}

func (r *ClientCache) Get(v string) (*Client, error) {

	val, err := r.v.Get(string(net.ParseIP(v).String()))
	if err != nil {
		return nil, err
	}

	cli, ok := val.(*Client)

	if !ok || cli == nil {
		return nil, ErrNotFound
	}
	return cli, nil
}

func (r *ClientCache) Set(k string, v *Client) {
	r.v.Set(string(net.ParseIP(k).String()), v)
}
