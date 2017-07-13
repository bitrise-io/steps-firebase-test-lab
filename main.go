package main

import (
	"fmt"
	"os"
	"encoding/base64"
	"io/ioutil"
	"encoding/json"
	"path"
	"github.com/kballard/go-shellquote"
	"os/exec"
	"errors"
	. "github.com/bootstraponline/bitrise-step-firebase-test-lab/utils"
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
		json.Unmarshal([]byte(gcloud_key), &parsedKeyFile)

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

	gcloud_options_value, err := GetRequiredEnv(GCLOUD_OPTIONS)
	if err != nil {
		return empty, err
	}

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

func executeGcloud(config *FirebaseConfig) ([]string, error) {
	if !config.Debug {
		RunCommand("gcloud config set project " + config.Project)
		RunCommand("gcloud auth activate-service-account --key-file " + config.KeyPath + " " + config.User)
	}

	// https://cloud.google.com/sdk/gcloud/reference/firebase/test/android/run
	gcloudOptions, err := shellquote.Split(config.Options)
	if err != nil {
		return make([]string, 0), err
	}
	fmt.Println("user options: ", gcloudOptions)

	// TODO: skip setting by default if these flags were specified by the user
	// Set --app, --test, --results-bucket, --results-dir and test type
	args := make([]string, 0)

	if IsEmpty(config.TestApk) {
		args = append(args, "robo")
	} else {
		args = append(args, "instrumentation")
		args = append(args, "--test", config.TestApk)
	}

	args = append(args, "--app", config.AppApk)
	args = append(args, "--results-bucket="+config.ResultsBucket)
	gcs_object := NewGcsObjectName()
	args = append(args, "--results-dir="+gcs_object)

	fmt.Println("auto options: ", args)

	exportGcsDir(config.ResultsBucket, gcs_object)

	return args, nil
}

func main() {
	config, err := NewFirebaseConfig()
	FatalError(err)

	// todo: pass string slice to run command
	_, err = executeGcloud(config)
	FatalError(err)
	os.Exit(0)
}
