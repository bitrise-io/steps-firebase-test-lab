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

type gcloudKeyFile struct {
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

	gcloudUser := getOptionalEnv(gcloudUser)
	gcloudProject := getOptionalEnv(gcloudProject)

	appApk, err := getRequiredEnv(appApk)
	if err != nil {
		return empty, err
	}

	err = fileExists(appApk)
	if err != nil {
		return empty, err
	}

	testApk := getOptionalEnv(testApk)
	if !isEmpty(testApk) {
		err = fileExists(testApk)
		if err != nil {
			return empty, err
		}
	}

	gcloudKeyBase64, err := getRequiredEnv(gcloudKey)
	if err != nil {
		return empty, err
	}

	gcloudKey, err := base64.StdEncoding.DecodeString(gcloudKeyBase64)
	if err != nil {
		return empty, err
	}

	emptyGcloudUser := isEmpty(gcloudUser)
	emptyGcloudProject := isEmpty(gcloudProject)

	if emptyGcloudUser || emptyGcloudProject {
		parsedKeyFile := gcloudKeyFile{}
		err = json.Unmarshal([]byte(gcloudKey), &parsedKeyFile)
		if err != nil {
			return empty, err
		}

		if emptyGcloudUser {
			gcloudUser = parsedKeyFile.ClientEmail
			if isEmpty(gcloudUser) {
				return empty, errors.New("gcloudUser not defined in env or gcloud key")

			}
		}

		if emptyGcloudProject {
			gcloudProject = parsedKeyFile.ProjectID
			if isEmpty(gcloudProject) {
				return empty, errors.New("gcloudProject not defined in env or gcloud key")
			}
		}
	}

	homeDir, err := getRequiredEnv(home)
	if err != nil {
		return empty, err
	}

	keyFilePath := path.Join(homeDir, "gcloudkey.json")
	err = ioutil.WriteFile(keyFilePath, gcloudKey, 0644)
	if err != nil {
		return empty, err
	}

	gcloudBucketValue, err := getRequiredEnv(gcloudBucket)
	if err != nil {
		return empty, err
	}

	gcloudOptionsValue := getOptionalEnv(gcloudOptions)

	return &firebaseConfig{
		ResultsBucket: gcloudBucketValue,
		User:          gcloudUser,
		Project:       gcloudProject,
		KeyPath:       keyFilePath,
		AppApk:        appApk,
		TestApk:       testApk,
		Options:       gcloudOptionsValue,
		Debug:         false,
	}, nil
}

func exportGcsDir(bucket string, object string) error {
	gcsResultsDir := "gs://" + bucket + "/" + object
	fmt.Println("Exporting ", gcsResultsDir, " ", gcsResultsDir)
	cmdLog, err := exec.Command("bitrise", "envman", "add", "--key", gcsResultsDir, "--value", gcsResultsDir).CombinedOutput()
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
