package e2e

import (
	lib "github.com/bazooka-ci/bazooka/commons"

	"github.com/stretchr/testify/require"

	"testing"
	"time"
)

func TestSimpleGoProject(t *testing.T) {
	bzk := NewBazooka(t)
	defer bzk.Teardown()

	repo := bzk.NewRepository()
	repo.ImportDir("data/go-project")
	repo.GitAddAll()
	repo.GitCommit("Point of inception")

	proj, err := bzk.Api.Project.Create("goproj", "git", repo.CloneURL())

	require.NoError(t, err, "error while creating a project")
	t.Logf("Created project: %v", proj)

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

func TestSimpleJavaProject(t *testing.T) {
	bzk := NewBazooka(t)
	defer bzk.Teardown()

	repo := bzk.NewRepository()
	repo.ImportDir("data/java-project")
	repo.GitAddAll()
	repo.GitCommit("Point of inception")

	proj, err := bzk.Api.Project.Create("goproj", "git", repo.CloneURL())

	require.NoError(t, err, "error while creating a project")
	t.Logf("Created project: %v", proj)

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

func TestSimplePythonProject(t *testing.T) {
	bzk := NewBazooka(t)
	defer bzk.Teardown()

	repo := bzk.NewRepository()
	repo.ImportDir("data/python-project")
	repo.GitAddAll()
	repo.GitCommit("Point of inception")

	proj, err := bzk.Api.Project.Create("goproj", "git", repo.CloneURL())

	require.NoError(t, err, "error while creating a project")
	t.Logf("Created project: %v", proj)

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

func TestSimpleNodejsProject(t *testing.T) {
	bzk := NewBazooka(t)
	defer bzk.Teardown()

	repo := bzk.NewRepository()
	repo.ImportDir("data/nodejs-project")
	repo.GitAddAll()
	repo.GitCommit("Point of inception")

	proj, err := bzk.Api.Project.Create("goproj", "git", repo.CloneURL())

	require.NoError(t, err, "error while creating a project")
	t.Logf("Created project: %v", proj)

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
