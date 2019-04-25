package discover

import (
	"context"
	"github.com/go-kit/kit/sd/etcdv3"
	"time"
)

func RegesiterEtcd(etcdv3Addr string) (etcdv3client etcdv3.Client, err error) {
	
	etcdConfig := etcdv3.ClientOptions{
		DialTimeout:   time.Second * 3,
		DialKeepAlive: time.Second * 3,
	}
	
	return etcdv3.NewClient(context.Background(), []string{etcdv3Addr}, etcdConfig)
}
