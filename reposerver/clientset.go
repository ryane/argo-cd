package reposerver

import (
	"crypto/tls"
	"time"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/argoproj/argo-cd/reposerver/repository"
	"github.com/argoproj/argo-cd/util"
)

// Clientset represets repository server api clients
type Clientset interface {
	NewRepoServerClient() (util.Closer, repository.RepoServerServiceClient, error)
}

type clientSet struct {
	address string
}

func (c *clientSet) NewRepoServerClient() (util.Closer, repository.RepoServerServiceClient, error) {
	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithMax(3),
		grpc_retry.WithBackoff(grpc_retry.BackoffLinear(1000 * time.Millisecond)),
	}
	conn, err := grpc.Dial(c.address,
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})),
		grpc.WithStreamInterceptor(grpc_retry.StreamClientInterceptor(retryOpts...)),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(retryOpts...)))
	if err != nil {
		log.Errorf("Unable to connect to repository service with address %s", c.address)
		return nil, nil, err
	}
	return conn, repository.NewRepoServerServiceClient(conn), nil
}

// NewRepoServerClientset creates new instance of repo server Clientset
func NewRepoServerClientset(address string) Clientset {
	return &clientSet{address: address}
}
