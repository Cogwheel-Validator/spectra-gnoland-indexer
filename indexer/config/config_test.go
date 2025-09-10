package config_test

import (
	"testing"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/config"
)

func TestNormalLoadConfig(t *testing.T) {
	conf, err := config.LoadConfig("testdata/test.yml")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	t.Log(conf.RpcUrl)
	t.Log(conf.PoolMaxConns)
	t.Log(conf.PoolMinConns)
	t.Log(conf.PoolMaxConnLifetime)
	t.Log(conf.PoolMaxConnIdleTime)
	t.Log(conf.PoolHealthCheckPeriod)
	t.Log(conf.PoolMaxConnLifetimeJitter)
	t.Log(conf.LivePooling)
	t.Log(conf.MaxBlockChunkSize)
	t.Log(conf.MaxTransactionChunkSize)
	t.Log(conf.ChainName)
}

func TestErrorLoadConfig(t *testing.T) {
	// this test should fail because the config is invalid
	_, err := config.LoadConfig("testdata/text2.yml")
	if err != nil {
		t.Log("This is expected because the rpc url is invalid")
		t.Skipf("failed to load config: %v", err)
	}

	_, err = config.LoadConfig("testdata/test3.yml")
	if err != nil {
		t.Log("This is expected because the config is invalid")
		t.Skipf("failed to load config: %v", err)
	}
}
