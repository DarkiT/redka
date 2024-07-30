package redka

import (
	"net"
	"strconv"
	"strings"

	"github.com/darkit/redka"
	"github.com/darkit/redka/internal/server"
	"github.com/darkit/slog"
	_ "modernc.org/sqlite"
)

const (
	servicePort = 6379
	memoryURI   = "file:/data.db?vfs=memdb"
)

type config struct {
	Host    string
	Port    int
	Path    string
	Verbose bool

	svr    *server.Server
	logger *slog.Logger
}

func NewService(host string, port int, dbPath ...string) (*server.Server, error) {
	c := &config{}

	if len(dbPath) == 0 {
		c.Path = memoryURI
	} else {
		c.Path = dbPath[0]
	}

	c.Host = host

	if 1 <= port || port >= 65535 {
		c.Port = 0
	} else {
		c.Port = port
	}

	opts := redka.Options{
		DriverName: "sqlite",
		Logger:     slog.Default(),
		Pragma:     map[string]string{},
	}

	db, err := redka.Open(c.Path, &opts)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	c.svr = server.New(c.Addr(), db)
	c.logger = opts.Logger

	return c.svr, nil
}

func (c *config) GetRedka() *server.Server {
	if c.svr == nil {
		service, err := NewService("", servicePort)
		if err != nil {
			c.logger.Errorf("NewService error: %s", err.Error())
			return nil
		}
		c.svr = service
	}
	return c.svr
}

func (c *config) Start() {
	c.svr.Start()
}

func (c *config) Stop() error {
	return c.svr.Stop()
}

func (c *config) Addr() string {
	if c.Host == "" {
		c.Host = getlocalIP()
	}
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}

func getlocalIP() (ip string) {
	conn, err := net.Dial("udp", "119.29.29.29:53")
	if err != nil {
		return
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ip = strings.Split(localAddr.String(), ":")[0]
	return
}
