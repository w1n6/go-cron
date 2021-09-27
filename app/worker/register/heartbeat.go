package register

import (
	"context"
	"go-cron/app/common"
	"strconv"
	"time"

	"github.com/coreos/etcd/clientv3"
)

//健康状态
type HealthStatus int

const (
	Normal      = iota //节点正常
	Delay              //有延迟
	Unavailable        //不可用
)

//HeartBeat 定义了etcd连接，cancelFunc用于停止心跳，本机状态，注册路径
type HeartBeat struct {
	client  *clientv3.Client //etcd连接
	kv      clientv3.KV
	lease   clientv3.Lease
	leaseID clientv3.LeaseID

	isLive     bool               //是否存活
	cancelCtx  context.Context    //停止上下文
	cancelFunc context.CancelFunc //用于停止心跳
	status     int                //本机状态
	regkey     string             //注册路径
}

var hb *HeartBeat

//初始化函数(初始化etcd连接，获取regKey)
func HeartBeatInit() error {
	var (
		conf   common.Config
		client *clientv3.Client
		kv     clientv3.KV
		lease  clientv3.Lease
		ip     string
		err    error
	)
	//获取配置
	conf = *common.GetConfig()

	//初始化etcd连接
	if client, err = clientv3.New(clientv3.Config{
		Endpoints:   conf.Endpoints,
		DialTimeout: time.Millisecond * time.Duration(conf.DialTimeout),
	}); err != nil {
		return err
	}

	//初始化KV Lease
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)

	//初始化本机信息
	if ip, err = getLocalIP(); err != nil {
		return err
	}
	hb = &HeartBeat{
		client: client,
		kv:     kv,
		lease:  lease,
		isLive: false,
		status: Unavailable,
		regkey: common.JobWorkerDir + ip,
	}
	return nil
}

//获取全局HeartBeat
func GetHeartBeat() *HeartBeat {
	if hb == nil {
		HeartBeatInit()
		return hb
	}
	return hb
}

//设置健康状态
func SetHealth(status int) {
	hb := GetHeartBeat()
	//设置状态
	hb.status = status

	//暂停后自动恢复心跳
	hb.Stop()
}

//心跳初始化
func (hb *HeartBeat) KeepAlive() {
	var (
		leaseGrantResp *clientv3.LeaseGrantResponse
		keepAliveChan  <-chan *clientv3.LeaseKeepAliveResponse
		keepAliveResp  *clientv3.LeaseKeepAliveResponse
		retry          func()
		err            error
	)

	//过一秒重试
	retry = func() {
		hb.Stop()
		time.Sleep(1 * time.Second)
	}
	//keepalive
	go func() {
		for {
			// 创建租约
			if leaseGrantResp, err = hb.lease.Grant(context.TODO(), 10); err != nil {
				retry()
			}

			//获取leaseId
			hb.leaseID = leaseGrantResp.ID

			//worker存活
			hb.isLive = true

			// 自动续租
			if keepAliveChan, err = hb.lease.KeepAlive(context.TODO(), hb.leaseID); err != nil {
				retry()
			}

			hb.cancelCtx, hb.cancelFunc = context.WithCancel(context.TODO())

			//注册到etcd
			if _, err = hb.kv.Put(hb.cancelCtx, hb.regkey, strconv.Itoa(hb.status), clientv3.WithLease(hb.leaseID)); err != nil {
				retry()
			}

			// 处理续租应答
			for {
				keepAliveResp = <-keepAliveChan
				if keepAliveResp == nil {
					retry()
				}
			}
		}
	}()
}

//停止心跳
func (hb *HeartBeat) Stop() {
	//销毁key
	if hb.cancelFunc != nil {
		hb.cancelFunc()
	}
	//释放租约
	if hb.isLive {
		hb.lease.Revoke(context.TODO(), hb.leaseID)
	}
}
