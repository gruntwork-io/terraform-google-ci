package test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/gcp"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/require"
)

// The following test ensures the Cloud Build / GitHub example works as expected. It performs the
// following test logic:
//
// 1. Clones a Docker Sample App to its own folder.
// 2. Deploys all of the example resources using Terraform.
// 3. Waits for the GKE nodes to be available.
// 5. Commits a test file and pushes it to trigger a build.
// 6. Polls the Cloud Build API to check the build was successful.
// 7. Cleans up all of the resources using Terraform destroy.
// 8. Ensures all of the GCR images are removed.
func TestCloudBuildGitHubGke(t *testing.T) {
	t.Parallel()

	// Uncomment any of the following to skip that section during the test
	//os.Setenv("SKIP_create_test_copy_of_examples", "true")
	//os.Setenv("SKIP_create_terratest_options", "true")
	//os.Setenv("SKIP_terraform_apply", "true")
	//os.Setenv("SKIP_configure_kubectl", "true")
	//os.Setenv("SKIP_wait_for_workers", "true")
	//os.Setenv("SKIP_trigger_build", "true")
	//os.Setenv("SKIP_wait_for_build", "true")
	//os.Setenv("SKIP_cleanup", "true")

	// Create a directory path that won't conflict
	workingDir := filepath.Join(".", "stages", t.Name())

	test_structure.RunTestStage(t, "create_test_copy_of_examples", func() {
		testFolder := test_structure.CopyTerraformFolderToTemp(t, "..", "examples")
		logger.Logf(t, "path to test folder %s\n", testFolder)
		terraformModulePath := filepath.Join(testFolder, "cloud-build-github-gke")
		test_structure.SaveString(t, workingDir, "cloudBuildGitHubGkeTerraformModulePath", terraformModulePath)
	})

	test_structure.RunTestStage(t, "create_terratest_options", func() {
		cloudBuildGitHubGkeTerraformModulePath := test_structure.LoadString(t, workingDir, "cloudBuildGitHubGkeTerraformModulePath")
		tmpKubeConfigPath := k8s.CopyHomeKubeConfigToTemp(t)
		kubectlOptions := k8s.NewKubectlOptions("", tmpKubeConfigPath, "")
		uniqueID := random.UniqueId()
		project := gcp.GetGoogleProjectIDFromEnvVar(t)
		region := gcp.GetRandomRegion(t, project, nil, nil)
		githubOrg := "gruntwork-io"
		githubRepo := "sample-app-docker"
		gkeClusterTerratestOptions := createTestGitHubTerraformOptions(t, uniqueID, project, region, githubOrg, githubRepo, cloudBuildGitHubGkeTerraformModulePath)
		test_structure.SaveString(t, workingDir, "uniqueID", uniqueID)
		test_structure.SaveString(t, workingDir, "project", project)
		test_structure.SaveString(t, workingDir, "region", region)
		test_structure.SaveTerraformOptions(t, workingDir, gkeClusterTerratestOptions)
		test_structure.SaveKubectlOptions(t, workingDir, kubectlOptions)
	})

	defer test_structure.RunTestStage(t, "cleanup", func() {
		project := test_structure.LoadString(t, workingDir, "project")
		buildID := test_structure.LoadString(t, workingDir, "buildID")

		build := gcp.GetBuild(t, project, buildID)
		for _, image := range build.GetImages() {
			gcp.DeleteGCRRepo(t, image)
		}

		gkeClusterTerratestOptions := test_structure.LoadTerraformOptions(t, workingDir)
		terraform.Destroy(t, gkeClusterTerratestOptions)

		kubectlOptions := test_structure.LoadKubectlOptions(t, workingDir)
		err := os.Remove(kubectlOptions.ConfigPath)
		require.NoError(t, err)
	})

	test_structure.RunTestStage(t, "terraform_apply", func() {
		gkeClusterTerratestOptions := test_structure.LoadTerraformOptions(t, workingDir)
		terraform.InitAndApply(t, gkeClusterTerratestOptions)
	})

	test_structure.RunTestStage(t, "configure_kubectl", func() {
		gkeClusterTerratestOptions := test_structure.LoadTerraformOptions(t, workingDir)
		kubectlOptions := test_structure.LoadKubectlOptions(t, workingDir)
		project := test_structure.LoadString(t, workingDir, "project")
		region := test_structure.LoadString(t, workingDir, "region")
		clusterName := gkeClusterTerratestOptions.Vars["cluster_name"].(string)

		// gcloud beta container clusters get-credentials example-cluster --region australia-southeast1 --project dev-sandbox-123456
		cmd := shell.Command{
			Command: "gcloud",
			Args:    []string{"beta", "container", "clusters", "get-credentials", clusterName, "--region", region, "--project", project},
			Env: map[string]string{
				"KUBECONFIG": kubectlOptions.ConfigPath,
			},
		}

		shell.RunCommand(t, cmd)
	})

	test_structure.RunTestStage(t, "wait_for_workers", func() {
		kubectlOptions := test_structure.LoadKubectlOptions(t, workingDir)
		verifyGkeNodesAreReady(t, kubectlOptions)
	})

	test_structure.RunTestStage(t, "trigger_build", func() {
		// write a test file
		date := []byte(fmt.Sprintf("%s\n", time.Now().String()))
		testFile := fmt.Sprintf("%s/auto-committed.txt", os.Getenv("SAMPLE_APP_DIR"))
		err := ioutil.WriteFile(testFile, date, 0644)
		if err != nil {
			t.Fatalf("Could not write temporary file")
		}

		// commit and push
		cmd := shell.Command{
			Command:    "git-add-commit-push",
			Args:       []string{"--remote-name", "origin", "--path", testFile, "--message", "triggering a build", "--skip-git-pull", "--skip-ci-flag", ""},
			WorkingDir: os.Getenv("SAMPLE_APP_DIR"),
		}

		shell.RunCommand(t, cmd)
	})

	test_structure.RunTestStage(t, "wait_for_build", func() {
		gkeClusterTerratestOptions := test_structure.LoadTerraformOptions(t, workingDir)
		triggerID := terraform.Output(t, gkeClusterTerratestOptions, "trigger_id")
		project := test_structure.LoadString(t, workingDir, "project")
		buildID := verifyBuildWasSuccessful(t, project, triggerID)
		test_structure.SaveString(t, workingDir, "buildID", buildID)
	})
}
