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

type GcloudKeyFile struct {
	ProjectID   string `json:"project_id"`
	ClientEmail string `json:"client_email"`
}

type FirebaseConfig struct {
	ResultsBucket string
	Options       string
	User          string
	Project       string
	KeyPath       string
	AppApk        string
	TestApk       string
	Debug         bool
}

func NewFirebaseConfig() (*FirebaseConfig, error) {
	empty := &FirebaseConfig{}

	gcloud_user := GetOptionalEnv(GCLOUD_USER)
	gcloud_project := GetOptionalEnv(GCLOUD_PROJECT)

	app_apk, err := GetRequiredEnv(APP_APK)
	if err != nil {
		return empty, err
	}

	err = FileExists(app_apk)
	if err != nil {
		return empty, err
	}

	test_apk := GetOptionalEnv(TEST_APK)
	if !IsEmpty(test_apk) {
		err = FileExists(test_apk)
		if err != nil {
			return empty, err
		}
	}

	gcloud_key_base64, err := GetRequiredEnv(GCLOUD_KEY)
	if err != nil {
		return empty, err
	}

	gcloud_key, err := base64.StdEncoding.DecodeString(gcloud_key_base64)
	if err != nil {
		return empty, err
	}

	empty_gcloud_user := IsEmpty(gcloud_user)
	empty_gcloud_project := IsEmpty(gcloud_project)

	if empty_gcloud_user || empty_gcloud_project {
		parsedKeyFile := GcloudKeyFile{}
		err = json.Unmarshal([]byte(gcloud_key), &parsedKeyFile)
		if err != nil {
			return empty, err
		}

		if empty_gcloud_user {
			gcloud_user = parsedKeyFile.ClientEmail
			if IsEmpty(gcloud_user) {
				return empty, errors.New("GCLOUD_USER not defined in env or gcloud key")

			}
		}

		if empty_gcloud_project {
			gcloud_project = parsedKeyFile.ProjectID
			if IsEmpty(gcloud_project) {
				return empty, errors.New("GCLOUD_PROJECT not defined in env or gcloud key")
			}
		}
	}

	home_dir, err := GetRequiredEnv(HOME)
	if err != nil {
		return empty, err
	}

	key_file_path := path.Join(home_dir, "gcloudkey.json")
	err = ioutil.WriteFile(key_file_path, gcloud_key, 0644)
	if err != nil {
		return empty, err
	}

	gcloud_bucket_value, err := GetRequiredEnv(GCLOUD_BUCKET)
	if err != nil {
		return empty, err
	}

	gcloud_options_value := GetOptionalEnv(GCLOUD_OPTIONS)

	return &FirebaseConfig{
		ResultsBucket: gcloud_bucket_value,
		User:          gcloud_user,
		Project:       gcloud_project,
		KeyPath:       key_file_path,
		AppApk:        app_apk,
		TestApk:       test_apk,
		Options:       gcloud_options_value,
		Debug:         false,
	}, nil
}

func exportGcsDir(bucket string, object string) error {
	gcs_results_dir := "gs://" + bucket + "/" + object
	fmt.Println("Exporting ", GCS_RESULTS_DIR, " ", gcs_results_dir)
	cmdLog, err := exec.Command("bitrise", "envman", "add", "--key", GCS_RESULTS_DIR, "--value", gcs_results_dir).CombinedOutput()
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to export "+GCS_RESULTS_DIR+", error: %#v | output: %s", err.Error(), cmdLog))
	}

	return nil
}

func buildGcloudCommand(config *FirebaseConfig, gcs_object string) ([]string, error) {
	empty := make([]string, 0)
	if !config.Debug {
		err := RunCommand("gcloud config set project " + config.Project)
		if err != nil {
			return empty, err
		}

		err = RunCommand("gcloud auth activate-service-account --key-file " + config.KeyPath + " " + config.User)
		if err != nil {
			return empty, err
		}
	}

	// https://cloud.google.com/sdk/gcloud/reference/firebase/test/android/run
	userOptionsSlice, err := shellquote.Split(config.Options)
	if err != nil {
		return empty, err
	}

	userOptionsSet := GcloudOptionsToSet(userOptionsSlice)

	// Set --app, --test, --results-bucket, --results-dir and test type
	// Use user values for flags if supplied.
	args := make([]string, 0)
	args = append(args, "gcloud", "firebase", "test", "android", "run")

	const TYPE_FLAG = "--type"
	const TEST_FLAG = "--test"
	const APP_FLAG = "--app"
	const RESULTS_BUCKET_FLAG = "--results-bucket="
	const RESULTS_DIR_FLAG = "--results-dir="

	if IsEmpty(config.TestApk) {
		args = append(args, TYPE_FLAG, "robo")
	} else {
		args = append(args, TYPE_FLAG, "instrumentation")
		if !userOptionsSet[TEST_FLAG] {
			args = append(args, "--test", config.TestApk)
		}
	}

	if !userOptionsSet[APP_FLAG] {
		args = append(args, APP_FLAG, config.AppApk)
	}
	if !userOptionsSet[RESULTS_BUCKET_FLAG] {
		args = append(args, RESULTS_BUCKET_FLAG+config.ResultsBucket)
	}
	if !userOptionsSet[RESULTS_DIR_FLAG] {
		args = append(args, RESULTS_DIR_FLAG+gcs_object)
	}

	// Don't export results bucket when it's user defined.
	if !userOptionsSet[RESULTS_BUCKET_FLAG] || !userOptionsSet[RESULTS_DIR_FLAG] {
		err = exportGcsDir(config.ResultsBucket, gcs_object)
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
	config, err := NewFirebaseConfig()
	FatalError(err)

	gcsCommand, err := buildGcloudCommand(config, NewGcsObjectName())
	FatalError(err)

	log.Printf(command.PrintableCommandArgs(false, gcsCommand))
	fmt.Println()

	err = RunCommandSlice(gcsCommand)
	FatalError(err)

	os.Exit(0)
}
