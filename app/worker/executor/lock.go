package executor

import (
	"context"
	"go-cron/app/common"

	"github.com/coreos/etcd/clientv3"
)

// 分布式锁(TXN事务)
type JobLock struct {
	// etcd客户端
	kv    clientv3.KV
	lease clientv3.Lease

	jobName    string             // 任务名
	cancelFunc context.CancelFunc // 用于终止自动续租
	leaseId    clientv3.LeaseID   // 租约ID
	isLocked   bool               // 是否上锁成功
}

func CreateLock(jobMgr *common.JobMgr, jobName string) *JobLock {
	jobLock := &JobLock{
		kv:    jobMgr.KV,
		lease: jobMgr.Lease,

		jobName: jobName,
	}
	return jobLock
}

func (lock *JobLock) Lock() error {
	var (
		leaseGrantResponse *clientv3.LeaseGrantResponse
		cancelCtx          context.Context
		cancelFunc         context.CancelFunc
		leaseID            clientv3.LeaseID
		keepRespChan       <-chan *clientv3.LeaseKeepAliveResponse
		key                string
		txn                clientv3.Txn
		txnResp            *clientv3.TxnResponse
		release            func()
		err                error
	)

	//创建租约 5秒
	if leaseGrantResponse, err = lock.lease.Grant(context.TODO(), 5); err != nil {
		lock.lease.Revoke(context.TODO(), leaseID)
		return err
	}

	//创建一个取消函数
	cancelCtx, cancelFunc = context.WithCancel(context.TODO())

	//租约ID
	leaseID = leaseGrantResponse.ID

	//释放函数
	release = func() {
		cancelFunc()                               // 取消自动续租
		lock.lease.Revoke(context.TODO(), leaseID) //  释放租约
	}

	//自动续租
	if keepRespChan, err = lock.lease.KeepAlive(cancelCtx, leaseID); err != nil {
		release()
		return err
	}

	//处理续租应答
	go func() {
		// for keepResp := range keepRespChan { //取自动续租的应答
		// 	if keepResp == nil {
		// 		break
		// 	}
		// }
		var (
			keepResp *clientv3.LeaseKeepAliveResponse
		)
		for {
			keepResp = <-keepRespChan
			if keepResp == nil {
				goto END
			}
		}
	END:
	}()

	//创建事务
	txn = lock.kv.Txn(context.TODO())
	key = common.JobLockDir + lock.jobName
	txn.If(clientv3.Compare(clientv3.CreateRevision(key), "=", 0)).
		Then(clientv3.OpPut(key, " ", clientv3.WithLease(leaseID))).
		Else(clientv3.OpGet(key))

	//提交事务
	if txnResp, err = txn.Commit(); err != nil {
		release()
		return err
	}

	//如果失败,锁被占用
	if !txnResp.Succeeded {
		release()
		return common.Err_Lock_Already_Required
	}

	//如果成功
	lock.leaseId = leaseID
	lock.cancelFunc = cancelFunc
	lock.isLocked = true
	return nil
}

// 释放锁
func (lock *JobLock) Unlock() {
	//如果上锁
	if lock.isLocked {
		lock.cancelFunc()                               //释放key
		lock.lease.Revoke(context.TODO(), lock.leaseId) //释放租约
	}
}
