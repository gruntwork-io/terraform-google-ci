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

func TestCloudBuildCsrGke(t *testing.T) {
	t.Parallel()

	// Uncomment any of the following to skip that section during the test
	//os.Setenv("SKIP_create_test_copy_of_examples", "true")
	//os.Setenv("SKIP_create_terratest_options", "true")
	//os.Setenv("SKIP_terraform_apply", "true")
	//os.Setenv("SKIP_configure_kubectl", "true")
	//os.Setenv("SKIP_wait_for_workers", "true")
	//os.Setenv("SKIP_cleanup", "true")

	// Create a directory path that won't conflict
	workingDir := filepath.Join(".", "stages", t.Name())

	test_structure.RunTestStage(t, "create_test_copy_of_examples", func() {
		testFolder := test_structure.CopyTerraformFolderToTemp(t, "..", "examples")
		logger.Logf(t, "path to test folder %s\n", testFolder)
		terraformModulePath := filepath.Join(testFolder, "cloud-build-csr-gke")
		test_structure.SaveString(t, workingDir, "cloudBuildCsrGkeTerraformModulePath", terraformModulePath)
	})

	test_structure.RunTestStage(t, "create_terratest_options", func() {
		cloudBuildCsrGkeTerraformModulePath := test_structure.LoadString(t, workingDir, "cloudBuildCsrGkeTerraformModulePath")
		tmpKubeConfigPath := k8s.CopyHomeKubeConfigToTemp(t)
		kubectlOptions := k8s.NewKubectlOptions("", tmpKubeConfigPath, "")
		uniqueID := random.UniqueId()
		project := gcp.GetGoogleProjectIDFromEnvVar(t)
		region := gcp.GetRandomRegion(t, project, nil, nil)
		gkeClusterTerratestOptions := createTestGKEClusterTerraformOptions(t, uniqueID, project, region, cloudBuildCsrGkeTerraformModulePath)
		test_structure.SaveString(t, workingDir, "uniqueID", uniqueID)
		test_structure.SaveString(t, workingDir, "project", project)
		test_structure.SaveString(t, workingDir, "region", region)
		test_structure.SaveTerraformOptions(t, workingDir, gkeClusterTerratestOptions)
		test_structure.SaveKubectlOptions(t, workingDir, kubectlOptions)
	})

	defer test_structure.RunTestStage(t, "cleanup", func() {
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

	// trigger a build
	test_structure.RunTestStage(t, "trigger_build", func() {
		gkeClusterTerratestOptions := test_structure.LoadTerraformOptions(t, workingDir)
		project := test_structure.LoadString(t, workingDir, "project")
		repoName := gkeClusterTerratestOptions.Vars["repository_name"].(string)

		// add the cloud source repository as a git remote
		// `git remote add google https://source.developers.google.com/p/[PROJECT_ID]/r/[REPO_NAME]`
		cmd := shell.Command{
			Command:    "git",
			Args:       []string{"remote", "add", "google", fmt.Sprintf("https://source.developers.google.com/p/%s/r/%s", project, repoName)},
			WorkingDir: os.Getenv("SAMPLE_APP_DIR"),
		}

		shell.RunCommand(t, cmd)

		// write a test file
		date := []byte(fmt.Sprintf("%s\n", time.Now().String()))
		testFile := fmt.Sprintf("%s/auto-committed.txt", os.Getenv("SAMPLE_APP_DIR"))
		err := ioutil.WriteFile(testFile, date, 0644)
		if err != nil {
			t.Fatalf("Could not write temporary file")
		}

		// commit and push
		cmd2 := shell.Command{
			Command:    "git-add-commit-push",
			Args:       []string{"--path", testFile, "--message", "triggering a build", "--skip-ci-flag", ""},
			WorkingDir: os.Getenv("SAMPLE_APP_DIR"),
		}

		shell.RunCommand(t, cmd2)
	})
}
