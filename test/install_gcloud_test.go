package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/stretchr/testify/require"
)

func TestInstallGCloud(t *testing.T) {
	t.Parallel()

	require.NoError(t, runInstallGCloudScript(t))
}

func runInstallGCloudScript(t *testing.T) error {
	cmd := shell.Command{Command: "../modules/gcp-helpers/bin/install-gcloud"}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	return err
}
