package redka

import (
	"errors"
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
	driverName  = "sqlite"
	memoryURI   = "file:/data.db?vfs=memdb"
)

type Server = server.Server

type config struct {
	Host    string
	Port    int
	Path    string
	Verbose bool

	svr    *Server
	db     *redka.DB
	logger *slog.Logger
}

func NewService(host string, port int, dbPath ...string) (svr *Server, err error) {
	c := &config{}

	if len(dbPath) == 0 {
		c.Path = memoryURI
	} else {
		c.Path = dbPath[0]
	}

	c.Host = host

	if 1 <= port && port >= 65535 {
		c.Port = 0
	} else {
		c.Port = port
	}

	opts := redka.Options{
		DriverName: driverName,
		Logger:     slog.Default(),
		Pragma:     map[string]string{},
	}

	slog.Infof("DbPath: %s", c.Path)
	c.db, err = redka.Open(c.Path, &opts)
	if err != nil {
		return nil, err
	}

	c.svr = server.New(c.Addr(), c.db)
	c.logger = opts.Logger

	c.Start()

	return c.svr, nil
}

func (c *config) GetRedka() *Server {
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

func (c *config) GetDb() (*redka.DB, error) {
	if c.db == nil {
		return nil, errors.New("redka not init")
	}
	return c.db, nil
}

func (c *config) Start() error {
	if c.db == nil {
		return errors.New("redka not init")
	}
	c.svr.Start()
	return nil
}

func (c *config) Stop() error {
	defer c.svr.Stop()
	return c.db.Close()
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
