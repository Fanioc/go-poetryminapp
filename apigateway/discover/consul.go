package discover

import (
	"github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
)

func ConnectConsul(consulAddr string) (etcdv3client consul.Client, err error) {
	consulConfig := api.DefaultConfig()
	consulConfig.Address = consulAddr
	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		return nil, err
	}
	return consul.NewClient(consulClient), nil
}
