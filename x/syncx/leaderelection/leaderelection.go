package leaderelection

import (
	"context"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type EctdLeaderElection struct {
	LeaderElectionConfig
	session  *concurrency.Session
	election *concurrency.Election
}

type LeaderElectionConfig struct {
	// Lock is the resource that will be used for locking
	Client *clientv3.Client

	// LeaseDuration is the duration that non-leader candidates will
	// wait to force acquire leadership. This is measured against time of
	// last observed ack.
	//
	// A client needs to wait a full LeaseDuration without observing a change to
	// the record before it can attempt to take over. When all clients are
	// shutdown and a new set of clients are started with different names against
	// the same leader record, they must wait the full LeaseDuration before
	// attempting to acquire the lease. Thus LeaseDuration should be as short as
	// possible (within your tolerance for clock skew rate) to avoid a possible
	// long waits in the scenario.
	//
	// Core clients default this value to 15 seconds.
	LeaseSeconds int

	// Callbacks are callbacks that are triggered during certain lifecycle
	// events of the LeaderElector
	Callbacks LeaderCallbacks

	Prefix string

	Identity string
}

type LeaderCallbacks struct {
	// OnStartedLeading is called when a LeaderElector client starts leading
	OnStartedLeading func(context.Context)
	// OnStoppedLeading is called when a LeaderElector client stops leading
	OnStoppedLeading func()
	// OnNewLeader is called when the client observes a leader that is
	// not the previously observed leader. This includes the first observed
	// leader when the client starts.
	OnNewLeader func(identity string)
}

func New(config LeaderElectionConfig) (*EctdLeaderElection, error) {
	session, err := concurrency.NewSession(config.Client, concurrency.WithTTL(config.LeaseSeconds))
	if err != nil {
		return nil, err
	}

	election := concurrency.NewElection(session, config.Prefix)

	return &EctdLeaderElection{
		LeaderElectionConfig: config,
		session:              session,
		election:             election,
	}, nil
}

func (le *EctdLeaderElection) Run(ctx context.Context) error {
	defer func() {
		le.Callbacks.OnStoppedLeading()
	}()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go le.observe(ctx)

	if err := le.election.Campaign(ctx, le.Identity); err != nil {
		return err
	}

	le.Callbacks.OnStartedLeading(ctx)

	return nil
}

func (le *EctdLeaderElection) observe(ctx context.Context) {
	if le.Callbacks.OnNewLeader == nil {
		return
	}

	ch := le.election.Observe(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case resp, ok := <-ch:
			if !ok {
				return
			}

			if len(resp.Kvs) == 0 {
				continue
			}

			leader := string(resp.Kvs[0].Value)
			if leader != le.Identity {
				go le.Callbacks.OnNewLeader(leader)
			}
		}
	}
}

func (le *EctdLeaderElection) Close() error {
	return le.session.Close()
}
