package etcd

import (
	"context"
	"go.etcd.io/etcd/client/v3"
	"log"
	"time"
)

type Register struct {
	etcdCli *clientv3.Client // etcd连接
	leaseId clientv3.LeaseID // 租约ID
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewEtcdRegister() (*Register, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"http://localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Printf("new etcd client failed,error %v \n", err)
		return nil, err
	}

	ctx, cancelFunc := context.WithCancel(context.Background())

	reg := &Register{
		etcdCli: client,
		ctx:     ctx,
		cancel:  cancelFunc,
	}
	return reg, nil
}

// CreateLease 创建租约
// expire 有效期/秒
func (r *Register) CreateLease(expire int64) error {
	res, err := r.etcdCli.Grant(r.ctx, expire)
	if err != nil {
		log.Printf("createLease failed,error %v \n", err)
		return err
	}

	r.leaseId = res.ID
	return nil
}

// BindLease 绑定租约
// 将租约和对应的KEY-VALUE绑定
func (r *Register) BindLease(key string, value string) error {
	res, err := r.etcdCli.Put(r.ctx, key, value, clientv3.WithLease(r.leaseId))
	if err != nil {
		log.Printf("bindLease failed,error %v \n", err)
		return err
	}

	log.Printf("bindLease success %v \n", res)
	return nil
}

// KeepAlive 续租
// 发送心跳，表明服务正常
func (r *Register) KeepAlive() (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	resChan, err := r.etcdCli.KeepAlive(r.ctx, r.leaseId)
	if err != nil {
		log.Printf("keepAlive failed,error %v \n", resChan)
		return resChan, err
	}

	return resChan, nil
}

// Watcher 监听
func (r *Register) Watcher(key string, resChan <-chan *clientv3.LeaseKeepAliveResponse) {
	for {
		select {
		case l := <-resChan:
			log.Printf("续约成功,val:%+v \n", l)
		case <-r.ctx.Done():
			log.Printf("续约关闭")
			return
		}
	}
}

func (r *Register) Close() error {
	r.cancel()

	log.Printf("closed...\n")

	// 撤销租约
	r.etcdCli.Revoke(r.ctx, r.leaseId)

	return r.etcdCli.Close()
}

// RegisterServer 注册服务
// expire 过期时间
func (r *Register) RegisterServer(serviceName, addr string, expire int64) (err error) {

	// 创建租约
	err = r.CreateLease(expire)
	if err != nil {
		return err
	}

	// 绑定租约
	err = r.BindLease(serviceName, addr)
	if err != nil {
		return err
	}

	// 续租
	keepAliveChan, err := r.KeepAlive()
	if err != nil {
		return err
	}

	// 监听续约
	go r.Watcher(serviceName, keepAliveChan)

	return nil
}
