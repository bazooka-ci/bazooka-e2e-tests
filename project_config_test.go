package e2e

import (
	"github.com/stretchr/testify/require"

	"testing"
)

const (
	projectConfigKey    = "a.b"
	projectConfigValue  = "Ze Value"
	projectConfigValue2 = "Anozer Value"
)

func TestProjectConfig(t *testing.T) {
	bzk := NewBazooka(t)
	defer bzk.Teardown()

	proj, err := bzk.Api.Project.Create("param-proj", "git", "nothing")
	require.NoError(t, err, "error while creating a project")
	t.Logf("Created project: %v", proj.ID)

	// set the first value
	err = bzk.Api.Project.Config.SetKey(proj.ID, projectConfigKey, projectConfigValue)
	require.NoError(t, err, "error while setting a project config key")

	// ensure it was stored
	require.Equal(t, projectConfigValue, getProjectConfigKey(bzk, proj.ID, projectConfigKey), "project config mismatch")

	// update the value
	err = bzk.Api.Project.Config.SetKey(proj.ID, projectConfigKey, projectConfigValue2)
	require.NoError(t, err, "error while setting a project config key")

	//ensure it was stored
	require.Equal(t, projectConfigValue2, getProjectConfigKey(bzk, proj.ID, projectConfigKey), "project config mismatch")

	// unset the value
	err = bzk.Api.Project.Config.UnsetKey(proj.ID, projectConfigKey)
	require.NoError(t, err, "error while setting a project config key")

	//ensure it was deleted
	require.Empty(t, getProjectConfigKey(bzk, proj.ID, projectConfigKey), "the key should have been deleted")

}

func getProjectConfigKey(bzk *Bzk, projectID, key string) string {
	cfg, err := bzk.Api.Project.Config.Get(projectID)
	require.NoError(bzk.t, err, "error while getting a project config key")
	return cfg[key]
}
