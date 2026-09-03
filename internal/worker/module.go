package worker

import (
	"github.com/userreksai/ecmdb-main/internal/worker/internal/job"
)

type Module struct {
	Svc Service
	Job *job.ServiceDiscoveryJob
}
