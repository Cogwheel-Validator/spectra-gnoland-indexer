package valkey

import (
	"context"
	"time"

	glide "github.com/valkey-io/valkey-glide/go/v2"
	glideCfg "github.com/valkey-io/valkey-glide/go/v2/config"
)

type ValkeyClient struct {
	client *glide.Client
}

func NewValkeyClient(host string, port int) (*ValkeyClient, error) {
	timeout := 5 * time.Second
	cfg := glideCfg.NewClientConfiguration()
	cfg.WithAddress(&glideCfg.NodeAddress{
		Host: host,
		Port: port,
	})
	cfg.WithConnectionTimeout(timeout)
	cfg.WithRequestTimeout(timeout)
	client, err := glide.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &ValkeyClient{
		client: client,
	}, nil
}

func (c *ValkeyClient) Close() {
	c.client.Close()
}

func (c *ValkeyClient) Increment(key string, ctx context.Context) (int64, error) {
	return c.client.Incr(ctx, key)
}

func (c *ValkeyClient) Expirer(key string, ctx context.Context, expiration time.Duration) (bool, error) {
	return c.client.Expire(ctx, key, expiration)
}
