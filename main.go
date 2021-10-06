package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/log"
	"github.com/kballard/go-shellquote"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
)

// GcloudKeyFile defines the project id & user
// must be exported for json.Unmarshal
type GcloudKeyFile struct {
	ProjectID   string `json:"project_id"`
	ClientEmail string `json:"client_email"`
}

type firebaseConfig struct {
	ResultsBucket string
	Options       string
	User          string
	Project       string
	KeyPath       string
	AppApk        string
	TestApk       string
	Debug         bool
}

func newFirebaseConfig() (*firebaseConfig, error) {
	empty := &firebaseConfig{}

	gcloudUserValue := getOptionalEnv(envKeyGcloudUser)
	gcloudProjectValue := getOptionalEnv(envKeyGcloudProject)

	appApkValue, err := getRequiredEnv(envKeyAppApk)
	if err != nil {
		return empty, err
	}

	err = fileExists(appApkValue)
	if err != nil {
		return empty, err
	}

	testApkValue := getOptionalEnv(envKeyTestApk)
	if !isEmpty(testApkValue) {
		err = fileExists(testApkValue)
		if err != nil {
			return empty, err
		}
	}

	gcloudKeyBase64, err := getRequiredEnv(envKeyGcloud)
	if err != nil {
		return empty, err
	}

	gcloudKey, err := base64.StdEncoding.DecodeString(gcloudKeyBase64)
	if err != nil {
		return empty, err
	}

	emptyGcloudUser := isEmpty(gcloudUserValue)
	emptyGcloudProject := isEmpty(gcloudProjectValue)

	if emptyGcloudUser || emptyGcloudProject {
		parsedKeyFile := GcloudKeyFile{}
		err = json.Unmarshal([]byte(gcloudKey), &parsedKeyFile)
		if err != nil {
			return empty, err
		}

		if emptyGcloudUser {
			gcloudUserValue = parsedKeyFile.ClientEmail
			if isEmpty(gcloudUserValue) {
				return empty, errors.New(envKeyGcloudUser + " not defined in env or gcloud key")

			}
		}

		if emptyGcloudProject {
			gcloudProjectValue = parsedKeyFile.ProjectID
			if isEmpty(gcloudProjectValue) {
				return empty, errors.New(envKeyGcloudProject + " not defined in env or gcloud key")
			}
		}
	}

	homeDir, err := getRequiredEnv(envKeyHome)
	if err != nil {
		return empty, err
	}

	keyFilePath := path.Join(homeDir, "gcloudkey.json")
	err = ioutil.WriteFile(keyFilePath, gcloudKey, 0644)
	if err != nil {
		return empty, err
	}

	gcloudBucketValue, err := getRequiredEnv(envKeyGcloudBucket)
	if err != nil {
		return empty, err
	}

	gcloudOptionsValue := getOptionalEnv(envKeyGcloudOptions)

	return &firebaseConfig{
		ResultsBucket: gcloudBucketValue,
		User:          gcloudUserValue,
		Project:       gcloudProjectValue,
		KeyPath:       keyFilePath,
		AppApk:        appApkValue,
		TestApk:       testApkValue,
		Options:       gcloudOptionsValue,
		Debug:         false,
	}, nil
}

func exportGcsDir(bucket string, object string) error {
    gcsResultsDirKey := "GCS_RESULTS_DIR"
	gcsResultsDir := "gs://" + bucket + "/" + object
	fmt.Println("Exporting ", gcsResultsDirKey, " ", gcsResultsDir)
	cmdLog, err := exec.Command("bitrise", "envman", "add", "--key", gcsResultsDirKey, "--value", gcsResultsDir).CombinedOutput()
	if err != nil {
		return fmt.Errorf("Failed to export "+gcsResultsDir+", error: %#v | output: %s", err.Error(), cmdLog)
	}

	return nil
}

func buildGcloudCommand(config *firebaseConfig, gcsObject string) ([]string, error) {
	empty := make([]string, 0)
	if !config.Debug {
		err := runCommand("gcloud config set project " + config.Project)
		if err != nil {
			return empty, err
		}

		err = runCommand("gcloud auth activate-service-account --key-file " + config.KeyPath + " " + config.User)
		if err != nil {
			return empty, err
		}
	}

	// https://cloud.google.com/sdk/gcloud/reference/firebase/test/android/run
	userOptionsSlice, err := shellquote.Split(config.Options)
	if err != nil {
		return empty, err
	}

	userOptionsSet := gcloudOptionsToSet(userOptionsSlice)

	// Set --app, --test, --results-bucket, --results-dir and test type
	// Use user values for flags if supplied.
	args := make([]string, 0)
	args = append(args, "gcloud", "firebase", "test", "android", "run")

	const TypeFlag = "--type"
	const TestFlag = "--test"
	const AppFlag = "--app"
	const ResultsBucketFlag = "--results-bucket="
	const ResultsDirFlag = "--results-dir="

	if isEmpty(config.TestApk) {
		args = append(args, TypeFlag, "robo")
	} else {
		args = append(args, TypeFlag, "instrumentation")
		if !userOptionsSet[TestFlag] {
			args = append(args, "--test", config.TestApk)
		}
	}

	if !userOptionsSet[AppFlag] {
		args = append(args, AppFlag, config.AppApk)
	}
	if !userOptionsSet[ResultsBucketFlag] {
		args = append(args, ResultsBucketFlag+config.ResultsBucket)
	}
	if !userOptionsSet[ResultsDirFlag] {
		args = append(args, ResultsDirFlag+gcsObject)
	}

	// Don't export results bucket when it's user defined.
	if !userOptionsSet[ResultsBucketFlag] || !userOptionsSet[ResultsDirFlag] {
		err = exportGcsDir(config.ResultsBucket, gcsObject)
		if err != nil {
			return empty, err
		}
	}

	if config.Debug {
		fmt.Println("auto args: ", args)
		fmt.Println("user args: ", userOptionsSlice)
	}

	return append(args, userOptionsSlice...), nil
}

func main() {
	config, err := newFirebaseConfig()
	fatalError(err)

	gcsCommand, err := buildGcloudCommand(config, newGcsObjectName())
	fatalError(err)

	log.Printf(command.PrintableCommandArgs(false, gcsCommand))
	fmt.Println()

	const InfrastructureFailure = 20
	const TryCount = 3

	// Note that gcloud CLI has a transparent retry of 3.
	// Retrying 3x here means we try up to 9 times in total.
	for i := 1; i <= TryCount; i++ {
		exitCode, err := runCommandSlice(gcsCommand)
		fatalError(err)

		if exitCode != InfrastructureFailure {
			break
		}
	}

	os.Exit(0)
}
