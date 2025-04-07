package register

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"time"

	"github.com/spf13/cast"
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	rootKey                 = "traefik"
	serviceKeyPrefixFmt     = rootKey + "/%s/services/%s/loadbalancer/"
	healthCheckKeyPrefixFmt = serviceKeyPrefixFmt + "healthcheck/"
	routerPrefixFmt         = "traefik/%s/routers/%s/"
)

var (
	UnsupportedErr = errors.New("unsupported this protocol")
)

type (
	Register interface {
		io.Closer
		Register(reader io.Reader) error
	}

	EtcdRegister struct {
		c *clientv3.Client
	}
)

func (e *EtcdRegister) Register(reader io.Reader) error {
	s, err := e.parseServiceConf(reader)
	if err != nil {
		return err
	}

	kvs, err := e.loadKV(s)
	if err != nil {
		return err
	}

	if err := e.register(kvs); err != nil {
		return err
	}
	slog.Info("register service success", slog.String("serverName", s.Name), slog.String("serviceAddr", s.Url))
	return nil
}

func (e *EtcdRegister) parseServiceConf(reader io.Reader) (*Service, error) {
	viper.SetConfigType("yaml")
	err := viper.ReadConfig(reader)
	if err != nil {
		return nil, err
	}
	s := &Service{}

	if err = viper.Unmarshal(s); err != nil {
		return nil, err
	}
	return s, nil
}

func (e *EtcdRegister) Close() error {
	return e.c.Close()
}

func (e *EtcdRegister) loadKV(s *Service) (map[string]string, error) {
	if !slices.Contains([]string{"http", "tcp", "upd"}, s.Protocol) {
		return nil, UnsupportedErr
	}
	res := make(map[string]string)
	// traefik/http/services/<service_name>/loadbalancer/servers/<n>/url
	res[fmt.Sprintf(serviceKeyPrefixFmt+"servers/%d/url", s.Protocol, s.Name, 0)] = s.Url
	// traefik/http/services/<service_name>/loadbalancer/healthcheck/interval
	res[fmt.Sprintf(healthCheckKeyPrefixFmt+"interval", s.Protocol, s.Name)] = s.HealthCheck.Interval
	// traefik/http/services/<service_name>/loadbalancer/healthcheck/method
	res[fmt.Sprintf(healthCheckKeyPrefixFmt+"method", s.Protocol, s.Name)] = s.HealthCheck.Method
	// traefik/http/services/<service_name>/loadbalancer/healthcheck/status
	res[fmt.Sprintf(healthCheckKeyPrefixFmt+"status", s.Protocol, s.Name)] = cast.ToString(s.HealthCheck.Status)
	// traefik/http/services/<service_name>/loadbalancer/healthcheck/path
	res[fmt.Sprintf(healthCheckKeyPrefixFmt+"path", s.Protocol, s.Name)] = s.HealthCheck.Path
	// traefik/http/services/<service_name>/loadbalancer/healthcheck/port
	res[fmt.Sprintf(healthCheckKeyPrefixFmt+"port", s.Protocol, s.Name)] = cast.ToString(s.HealthCheck.Port)
	// traefik/http/services/<service_name>/loadbalancer/healthcheck/timeout
	res[fmt.Sprintf(healthCheckKeyPrefixFmt+"timeout", s.Protocol, s.Name)] = s.HealthCheck.Timeout
	for _, r := range s.Routes {
		// traefik/http/routers/<router_name>/rule
		res[fmt.Sprintf(routerPrefixFmt+"rule", s.Protocol, r.Name)] = r.Rule
		for i, entry := range r.EntryPoints {
			// traefik/http/routers/<router_name>/entrypoints
			res[fmt.Sprintf(routerPrefixFmt+"entrypoints/%d", s.Protocol, r.Name, i)] = entry
		}
		// traefik/http/routers/<router_name>/service
		res[fmt.Sprintf(routerPrefixFmt+"service", s.Protocol, r.Name)] = s.Name
		// traefik/http/routers/<router_name>/priority
		res[fmt.Sprintf(routerPrefixFmt+"priority", s.Protocol, r.Name)] = cast.ToString(r.Priority)
		for i, mw := range r.Middleware {
			// traefik/http/routers/<router_name>/middlewares
			res[fmt.Sprintf(routerPrefixFmt+"middlewares/%d", s.Protocol, r.Name, i)] = mw
		}

	}
	return res, nil
}

func (e *EtcdRegister) register(kvs map[string]string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	for k, v := range kvs {
		_, err := e.c.Put(ctx, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewEtcdRegister(endpoints []string) (*EtcdRegister, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: time.Second * 5,
	})
	if err != nil {
		return nil, err
	}
	return &EtcdRegister{c: cli}, nil
}
