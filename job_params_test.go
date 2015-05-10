package e2e

import (
	lib "github.com/bazooka-ci/bazooka/commons"

	"github.com/stretchr/testify/require"

	"testing"
	"time"
)

func TestJobParameters(t *testing.T) {
	bzk := NewBazooka(t)
	defer bzk.Teardown()

	repo := bzk.NewRepository()
	repo.ImportDir("data/params-project")
	repo.GitAddAll()
	repo.GitCommit("Point of inception")

	proj, err := bzk.Api.Project.Create("param-proj", "git", repo.CloneURL())
	require.NoError(t, err, "error while creating a project")
	t.Logf("Created project: %v", proj.ID)

	job, err := bzk.Api.Project.StartJob(proj.ID, "master", []string{"PARAM=42"})
	require.NoError(t, err, "job creation failed")
	t.Logf("Started job: %v", job)

	jobStatus := bzk.WaitForJob(job.ID, 60*time.Second)

	require.Equal(t, lib.JOB_SUCCESS, jobStatus)
}

func TestJobParametersOverrideEnv(t *testing.T) {
	bzk := NewBazooka(t)
	defer bzk.Teardown()

	repo := bzk.NewRepository()
	repo.ImportDir("data/params-override-project")
	repo.GitAddAll()
	repo.GitCommit("Point of inception")

	proj, err := bzk.Api.Project.Create("param-proj", "git", repo.CloneURL())
	require.NoError(t, err, "error while creating a project")
	t.Logf("Created project: %v", proj.ID)

	job, err := bzk.Api.Project.StartJob(proj.ID, "master", []string{"PARAM=42"})
	require.NoError(t, err, "job creation failed")
	t.Logf("Started job: %v", job)

	jobStatus := bzk.WaitForJob(job.ID, 60*time.Second)

	require.Equal(t, lib.JOB_SUCCESS, jobStatus)
}
