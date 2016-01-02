package e2e

import (
	"time"

	"fmt"

	lib "github.com/bazooka-ci/bazooka/commons"
)

func (b *Bzk) WaitForJob(jobID string, timeoutAfter time.Duration) (lib.JobStatus, error) {
	b.t.Logf("Waiting for job %s", jobID)

	giveUp := time.After(timeoutAfter)
	start := time.Now()

	for {
		select {
		case <-time.After(500 * time.Millisecond):
			j, err := b.Api.Job.Get(jobID)
			switch {
			case err != nil:
				return "", fmt.Errorf("Error while getting the job %s status: %v", jobID, err)
			case j.Status != lib.JOB_RUNNING && j.Status != lib.JOB_PENDING:
				b.t.Logf("Job %s completed with status %v after %v", jobID, j.Status, time.Now().Sub(start))
				return j.Status, nil
			}

		case <-giveUp:
			return "", fmt.Errorf("Gave up waiting on job %s: didn't finish after %v", jobID, timeoutAfter)
		}
	}
}
