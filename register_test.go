package register

import (
	"bytes"
	_ "embed"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	//go:embed t.yml
	serviceCnf []byte
	r          *EtcdRegister
)

func TestMain(m *testing.M) {
	cli, err := NewEtcdRegister([]string{"192.168.31.43:12379", "192.168.31.43:22379", "192.168.31.43:32379"})
	if err != nil {
		panic(err)
	}
	r = cli
	m.Run()
	r.Close()
}

func TestRegister(t *testing.T) {
	s, err := r.parseServiceConf(bytes.NewReader(serviceCnf))
	assert.NoError(t, err)
	t.Logf("%+v", s)

	kv, err := r.loadKV(s)
	assert.NoError(t, err)
	t.Logf("%+v", kv)

	err = r.register(kv)
	assert.NoError(t, err)
}
