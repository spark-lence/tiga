package tiga

import (
	"context"
	"fmt"
	"path/filepath"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/mock/mockserver"
)

type EtcdDao struct {
	client *clientv3.Client
}

func NewEtcdDao(config *Configuration) *EtcdDao {
	enpoints := config.GetStrings("etcd.endpoints")
	timeout := config.GetDuration("etcd.timeout")
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   enpoints,
		DialTimeout: timeout,
	})
	if err != nil {
		panic(err)
	}
	return &EtcdDao{
		client: cli,
	}
}
func NewEtcdMockDao() *EtcdDao {
	servers, err := mockserver.StartMockServers(1)
	if err != nil {
		panic(err)
	}
	err = servers.StartAt(0)
	if err != nil {
		panic(err)
	}
	cli,err:=clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:0"},
	})
	if err!=nil{
		panic(err)
	}
	return &EtcdDao{client: cli}
}
func (e EtcdDao) Get(ctx context.Context, key string, opts ...clientv3.OpOption) ([]byte, error) {
	resp, err := e.client.Get(ctx, key, opts...)
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}
	return resp.Kvs[0].Value, nil
}
func (e EtcdDao) GetRnage(ctx context.Context, key string, opts ...clientv3.OpOption) ([]*mvccpb.KeyValue, error) {
	resp, err := e.client.Get(ctx, key, opts...)
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}
	return resp.Kvs, nil
}
func (e EtcdDao) GetString(ctx context.Context, key string) (string, error) {
	resp, err := e.Get(ctx, key)
	if err != nil {
		return "", err
	}
	return string(resp), nil
}
func (e EtcdDao) Put(ctx context.Context, key string, value string, opts ...clientv3.OpOption) error {
	_, err := e.client.Put(ctx, key, value, opts...)
	return err
}
func (e EtcdDao) PutSelect(ctx context.Context,
	prefixKey string,
	data map[string]interface{},
	selectFields []string,
	opts ...clientv3.OpOption) error {
	ops := make([]clientv3.Op, 0)
	for _, field := range selectFields {
		if val, ok := data[field]; ok {
			v, err := ValueToString(val)
			if err != nil {
				return err
			}
			ops = append(ops, clientv3.OpPut(filepath.Join(prefixKey, field), v))
		}
	}
	_, err := e.BatchOps(ctx, ops)
	return err
}
func (e EtcdDao) Delete(ctx context.Context, key string, opts ...clientv3.OpOption) error {
	_, err := e.client.Delete(ctx, key, opts...)
	return err
}
func (e EtcdDao) Watch(ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan {
	return e.client.Watch(ctx, key, opts...)
}
func (e EtcdDao) Close() error {
	return e.client.Close()
}
func (e EtcdDao) LeaseGrant(ctx context.Context, ttl int64) (*clientv3.LeaseGrantResponse, error) {
	return e.client.Grant(ctx, ttl)
}

func (e EtcdDao) LeaseKeepAlive(ctx context.Context, leaseID clientv3.LeaseID) (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	return e.client.KeepAlive(ctx, leaseID)
}
func (e EtcdDao) BatchOps(ctx context.Context, ops []clientv3.Op) (bool, error) {
	txn := e.client.Txn(context.Background())
	txnResp, err := txn.Then(ops...).Commit()
	if err != nil {
		return false, err
	}
	return txnResp.Succeeded, nil
}
func (e EtcdDao) BatchDelete(ctx context.Context, keys []string) (bool, error) {
	// ops := make([]clientv3.Op, 0)
	// for _, key := range keys {
	// 	ops = append(ops, clientv3.OpDelete(key))
	// }
	// return e.BatchPut(ctx, ops)
	return false, nil
}

func (e EtcdDao) BatchGet(ctx context.Context, ops []clientv3.Op) ([]*mvccpb.KeyValue, error) {
	resp,err:=e.client.Txn(ctx).Then(ops...).Commit()
	if err!=nil{
		return nil,err
	}
	if !resp.Succeeded{
		return nil,fmt.Errorf("batch get failed")
	}
	kvs:=make([]*mvccpb.KeyValue,0)
	for _,r:=range resp.Responses{
		kvs = append(kvs, r.GetResponseRange().Kvs...)
	}
	return kvs, nil
}
