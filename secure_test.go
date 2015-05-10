package e2e

import (
	lib "github.com/bazooka-ci/bazooka/commons"

	"github.com/stretchr/testify/require"

	"testing"
	"time"
)

const (
	sensitiveData = "ANSWER=42"
)

func TestSecureInEnv(t *testing.T) {
	bzk := NewBazooka(t)
	defer bzk.Teardown()

	repo := bzk.NewRepository()

	proj, err := bzk.Api.Project.Create("secure-proj", "git", repo.CloneURL())
	require.NoError(t, err, "error while creating a project")
	t.Logf("Created project: %v", proj)

	encryptedData, err := bzk.Api.Project.EncryptData(proj.ID, sensitiveData)
	require.NoError(t, err, "error while encrypting data")

	repo.ImportDir("data/secure-project")
	repo.Render(".bazooka.yml", map[string]interface{}{
		"Secure": encryptedData,
	})
	repo.GitAddAll()
	repo.GitCommit("Point of inception")

	job, err := bzk.Api.Project.StartJob(proj.ID, "master", nil)
	require.NoError(t, err, "job creation failed")
	t.Logf("Started job: %v", job)

	jobStatus := bzk.WaitForJob(job.ID, 60*time.Second)

	require.Equal(t, lib.JOB_SUCCESS, jobStatus)

	variants, err := bzk.Api.Job.Variants(job.ID)
	require.NoError(t, err, "error while listing job variants")

	require.Equal(t, 1, len(variants), "Should have exactly one variant")

	variant := variants[0]
	require.Equal(t, lib.JOB_SUCCESS, variant.Status)
}
