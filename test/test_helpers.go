package test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/gcp"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/stretchr/testify/assert"
	cloudbuildpb "google.golang.org/genproto/googleapis/devtools/cloudbuild/v1"
)

func verifyBuildWasSuccessful(t *testing.T, projectID string, triggerID string) string {
	statusMsg := fmt.Sprintf("Wait for build to complete.")
	retries := 30
	sleepBetweenRetries := 20 * time.Second

	successfulBuildID, err := retry.DoWithRetryE(
		t,
		statusMsg,
		retries,
		sleepBetweenRetries,
		func() (string, error) {
			builds := gcp.GetBuildsForTrigger(t, projectID, triggerID)

			if len(builds) == 0 {
				return "", errors.New("Build hasn't been triggered")
			}

			// assume the first build returned is the one we triggered.
			buildID := builds[0].GetId()
			build, err := gcp.GetBuildE(t, projectID, buildID)
			if err != nil {
				return "", err
			}

			if build.GetStatus() == cloudbuildpb.Build_QUEUED {
				return "", errors.New("Build is queued")
			}

			if build.GetStatus() == cloudbuildpb.Build_WORKING {
				return "", errors.New("Build is executing")
			}

			if build.GetStatus() != cloudbuildpb.Build_SUCCESS {
				return "", errors.New("Build is not successful")
			}

			return build.GetId(), nil
		},
	)
	if err != nil {
		logger.Logf(t, "Error waiting for the build to complete: %s", err)
		t.Fatal(err)
	}
	logger.Logf(t, "Build was successful")
	return successfulBuildID
}

// kubeWaitUntilNumNodes continuously polls the Kubernetes cluster until there are the expected number of nodes
// registered (regardless of readiness).
func kubeWaitUntilNumNodes(t *testing.T, kubectlOptions *k8s.KubectlOptions, numNodes int, retries int, sleepBetweenRetries time.Duration) {
	statusMsg := fmt.Sprintf("Wait for %d Kube Nodes to be registered.", numNodes)
	message, err := retry.DoWithRetryE(
		t,
		statusMsg,
		retries,
		sleepBetweenRetries,
		func() (string, error) {
			nodes, err := k8s.GetNodesE(t, kubectlOptions)
			if err != nil {
				return "", err
			}
			if len(nodes) != numNodes {
				return "", errors.New("Not enough nodes")
			}
			return "All nodes registered", nil
		},
	)
	if err != nil {
		logger.Logf(t, "Error waiting for expected number of nodes: %s", err)
		t.Fatal(err)
	}
	logger.Logf(t, message)
}

// Verify that all the nodes in the cluster reach the Ready state.
func verifyGkeNodesAreReady(t *testing.T, kubectlOptions *k8s.KubectlOptions) {
	kubeWaitUntilNumNodes(t, kubectlOptions, 3, 30, 10*time.Second)
	k8s.WaitUntilAllNodesReady(t, kubectlOptions, 30, 10*time.Second)
	readyNodes := k8s.GetReadyNodes(t, kubectlOptions)
	assert.Equal(t, len(readyNodes), 3)
}
