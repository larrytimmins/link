package locker

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Scalingo/go-utils/logger"
	"github.com/Scalingo/link/config"
	"github.com/Scalingo/link/models"
	"github.com/pkg/errors"
	"go.etcd.io/etcd/clientv3"
)

var (
	// ErrInvalidEtcdState is an error returned by IsMaster when the key supposed to contain the lock does not exist
	ErrInvalidEtcdState = errors.New("Invalid etcd state: key not found")
)

type etcdLocker struct {
	kvEtcd            clientv3.KV
	key               string
	config            config.Config
	ip                models.IP
	leaseManager      EtcdLeaseManager
	leaseSubscriberID string
	lock              *sync.Mutex
}

// NewEtcdLocker return an implemtation of Locker based on the ETCD database
func NewEtcdLocker(config config.Config, etcd *clientv3.Client, leaseManager EtcdLeaseManager, ip models.IP) Locker {
	key := fmt.Sprintf("%s/default/%s", models.ETCD_LINK_DIRECTORY, strings.Replace(ip.IP, "/", "_", -1))
	return &etcdLocker{
		kvEtcd:       etcd,
		key:          key,
		config:       config,
		ip:           ip,
		leaseManager: leaseManager,
		lock:         &sync.Mutex{},
	}
}

func (l *etcdLocker) Refresh(ctx context.Context) error {
	l.lock.Lock()
	defer l.lock.Unlock()
	log := logger.Get(ctx)

	// If we are not subscribed to lease changes yet
	if l.leaseSubscriberID == "" {
		id, err := l.leaseManager.SubscribeToLeaseChange(ctx, l.leaseChanged)
		if err != nil {
			log.WithError(err).Error("fail to subscribe to lease manager, will retry next time")
		} else {
			l.leaseSubscriberID = id
		}
	}

	leaseID, err := l.leaseManager.GetLease(ctx)
	if err != nil {
		return errors.Wrap(err, "fail to get lease ID")
	}

	// The goal of this transaction is to create the key with our leaseID only if this key does not exist
	// We use a transaction to make sure that concurrent tries wont interfere with each others.

	transactionTimeout := time.Duration(l.ip.KeepaliveInterval) * time.Second
	if transactionTimeout == 0 {
		transactionTimeout = l.config.KeepAliveInterval
	}
	transactionCtx, cancel := context.WithTimeout(ctx, transactionTimeout)
	defer cancel()

	_, err = l.kvEtcd.Txn(transactionCtx).
		// If the key does not exists (createRevision == 0)
		If(clientv3.Compare(clientv3.CreateRevision(l.key), "=", 0)).
		// Create it with our leaseID
		Then(clientv3.OpPut(l.key, l.config.Hostname, clientv3.WithLease(leaseID))).
		Commit()
	if err != nil {
		// We got an error. Notify the lease manager that there might be an issue and send the error.
		leaseManagerErr := l.leaseManager.MarkLeaseAsDirty(ctx, leaseID)
		if err != nil {
			log.WithError(leaseManagerErr).Error("fail to mark lease as dirty")
		}
		return errors.Wrapf(err, "fail to refresh lock")
	}

	return nil
}

func (l *etcdLocker) Unlock(ctx context.Context) error {
	_, err := l.kvEtcd.Delete(ctx, l.key)
	if err != nil {
		return errors.Wrap(err, "fail to unlock key")
	}
	return nil
}

func (l *etcdLocker) IsMaster(ctx context.Context) (bool, error) {
	// We do know that we are the master if:
	// - The key exist
	// - The lease associated to this key is our lease
	resp, err := l.kvEtcd.Get(ctx, l.key)
	if err != nil {
		return false, errors.Wrap(err, "fail to get lock")
	}

	// This could be the cause of a failure. If no key exist, this mean that all leases are expired
	// Or that all manager decided to stop managing this key
	if len(resp.Kvs) != 1 {
		return false, ErrInvalidEtcdState
	}

	leaseID, err := l.leaseManager.GetLease(ctx)
	if err != nil {
		return false, errors.Wrap(err, "fail to get current lease ID from manager")
	}

	return resp.Kvs[0].Lease == int64(leaseID), nil
}

func (l *etcdLocker) leaseChanged(ctx context.Context, oldLeaseID, newLeaseID clientv3.LeaseID) {
	log := logger.Get(ctx).WithFields(l.ip.ToLogrusFields())

	_, err := l.kvEtcd.Txn(ctx).
		// If the key does exists (createRevision != 0)
		If(clientv3.Compare(clientv3.CreateRevision(l.key), "!=", 0),
			// And we had the lock previously
			clientv3.Compare(clientv3.LeaseValue(l.key), "=", oldLeaseID)).
		// Replace it with the newLease
		Then(clientv3.OpPut(l.key, l.config.Hostname, clientv3.WithLease(newLeaseID))).
		Commit()

	if err != nil {
		log.WithError(err).Errorf("fail to change lease of key %s", l.key)
	}
}

// Stop will stop the lock we currently own. This will remove our lock if we are master and remove any subscription added to the lease manager
// If we fail to know if we are master or not, this will still try to delete the key (to prevent a situation where we could habe the key indefinitely)
// This is a failsafe since we should have called Unlock() a long time before calling this method
func (l *etcdLocker) Stop(ctx context.Context) error {
	log := logger.Get(ctx)
	log.Info("Stopping the locker")

	// First remove the subscription, if it fails: continue
	if l.leaseSubscriberID != "" {
		err := l.leaseManager.UnsubscribeToLeaseChange(ctx, l.leaseSubscriberID)
		if err != nil {
			log.WithError(err).Error("fail to remove subscription on lease changes")
		}
	}

	// Then check if we currently are master. (if there are any error: we are master!)
	isMaster, err := l.IsMaster(ctx)
	// Here the ErrInvalidEtcdState is expected if we are the last node: since no other node had taken this lock, the key is not found leading to this error
	if err != nil && err != ErrInvalidEtcdState {
		log.WithError(err).Error("We do not know if we are master or not. In doubt, delete lock. This may trigger a failover")
		isMaster = true
	}

	log.Info("We were master, deleting lock")
	if isMaster {
		// If we are the key master we should remove the key. Overwise since the lease is always
		// refreshed, we will be master forever.
		_, err := l.kvEtcd.Delete(ctx, l.key)
		if err != nil {
			return errors.Wrap(err, "fail to delete lock while stopping")
		}
	}

	return nil
}
