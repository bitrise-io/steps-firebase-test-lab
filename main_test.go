package main

import (
	"testing"
	"os"
	"io/ioutil"
	"github.com/stretchr/testify/assert"
	. "github.com/bootstraponline/bitrise-step-firebase-test-lab/utils"
	"errors"
)

// os.Exit(1) = test passes.

// GCLOUD_USER    // Optional. Read from keyfile
// GCLOUD_PROJECT // Optional. Read from keyfile
// GCLOUD_BUCKET  // Required
// GCLOUD_OPTIONS // Optional
// APP_APK        // Required
// TEST_APK       // Optional
// GCLOUD_KEY     // Required

const PATH = "PATH"

func resetEnv() {
	home := os.Getenv(HOME)
	path := os.Getenv(PATH)

	os.Clearenv()
	os.Setenv(HOME, home)
	os.Setenv(PATH, path)
}

func TestFileExists(t *testing.T) {
	assert := assert.New(t)

	err := FileExists("/tmp/nope.txt")
	assert.EqualError(err, "file doesn't exist: '/tmp/nope.txt'")

	err = FileExists("/tmp")
	assert.Equal(nil, err)
}

func TestRunCommand(t *testing.T) {
	assert := assert.New(t)

	err := RunCommand("true")
	assert.Equal(nil, err)
}

func TestGetRequiredEnv(t *testing.T) {
	assert := assert.New(t)

	const KEY = "KEY_THAT_IS_NOT_USED"
	const ENV_VALUE = "ENV_VALUE"
	_, err := GetRequiredEnv(KEY)
	assert.EqualError(err, KEY+" is not defined!")

	os.Setenv(KEY, ENV_VALUE)
	value, err := GetRequiredEnv(KEY)
	assert.Equal(ENV_VALUE, value)
}

func TestExecuteGcloud(t *testing.T) {
	assert := assert.New(t)
	gcloud_key, err := GetRequiredEnv(GCLOUD_KEY)
	assert.NoError(err)

	resetEnv()
	os.Setenv(GCLOUD_KEY, gcloud_key)

	os.Setenv(GCLOUD_BUCKET, "golang-bucket")
	os.Setenv(GCLOUD_OPTIONS, `--device-ids NexusLowRes
	 --os-version-ids 25 
	 --locales en 
	 --orientations portrait 
	 --timeout 25m 
	 --directories-to-pull=/sdcard 
	 --environment-variables ^:^coverage=true:coverageFile=/sdcard/coverage.ec`)

	app_apk_path := "/tmp/app.apk"
	test_apk_path := "/tmp/test.apk"

	if FileExists(app_apk_path) != nil {
		ioutil.WriteFile(app_apk_path, nil, 0644)
	}

	if FileExists(test_apk_path) != nil {
		ioutil.WriteFile(test_apk_path, nil, 0644)
	}

	os.Setenv(APP_APK, app_apk_path)
	os.Setenv(TEST_APK, test_apk_path)

	config, err := NewFirebaseConfig()
	config.Debug = true
	assert.NoError(err)

	gcs_object := NewGcsObjectName()
	result, err := executeGcloud(config, gcs_object)
	assert.NoError(err)

	assert.Equal([]string{
		"gcloud", "firebase", "test", "android", "run",
		"instrumentation",
		"--test", "/tmp/test.apk",
		"--app", "/tmp/app.apk",
		"--results-bucket=golang-bucket",
		"--results-dir=" + gcs_object,
		"--device-ids", "NexusLowRes",
		"--os-version-ids", "25",
		"--locales", "en",
		"--orientations", "portrait",
		"--timeout", "25m",
		"--directories-to-pull=/sdcard",
		"--environment-variables", "^:^coverage=true:coverageFile=/sdcard/coverage.ec",
	}, result)
}

func TestExecuteGcloudUserOverrides(t *testing.T) {
	assert := assert.New(t)
	gcloud_key, err := GetRequiredEnv(GCLOUD_KEY)
	assert.NoError(err)

	resetEnv()
	os.Setenv(GCLOUD_KEY, gcloud_key)

	// when user sets --test, --app, --results-bucket=, and --results-dir= then
	// we should use thoe values.
	os.Setenv(GCLOUD_BUCKET, "golang-bucket")
	os.Setenv(GCLOUD_OPTIONS, `
	 --test custom_test_apk
	 --app custom_app_apk
	 --results-bucket=custom_results_bucket
	 --results-dir=custom_results_dir
	 --device-ids NexusLowRes
	 --os-version-ids 25
	 --locales en
	 --orientations portrait
	 --timeout 25m
	 --directories-to-pull=/sdcard
	 --environment-variables ^:^coverage=true:coverageFile=/sdcard/coverage.ec`)

	app_apk_path := "/tmp/app.apk"
	test_apk_path := "/tmp/test.apk"

	if FileExists(app_apk_path) != nil {
		ioutil.WriteFile(app_apk_path, nil, 0644)
	}

	if FileExists(test_apk_path) != nil {
		ioutil.WriteFile(test_apk_path, nil, 0644)
	}

	os.Setenv(APP_APK, app_apk_path)
	os.Setenv(TEST_APK, test_apk_path)

	config, err := NewFirebaseConfig()
	config.Debug = true
	assert.NoError(err)

	gcs_object := NewGcsObjectName()
	result, err := executeGcloud(config, gcs_object)
	assert.NoError(err)

	assert.Equal([]string{
		"gcloud", "firebase", "test", "android", "run",
		"instrumentation",
		"--test", "custom_test_apk",
		"--app", "custom_app_apk",
		"--results-bucket=custom_results_bucket",
		"--results-dir=custom_results_dir",
		"--device-ids", "NexusLowRes",
		"--os-version-ids", "25",
		"--locales", "en",
		"--orientations", "portrait",
		"--timeout", "25m",
		"--directories-to-pull=/sdcard",
		"--environment-variables", "^:^coverage=true:coverageFile=/sdcard/coverage.ec",
	}, result)
}

func TestExecuteGcloudRobo(t *testing.T) {
	assert := assert.New(t)
	gcloud_key, err := GetRequiredEnv(GCLOUD_KEY)
	assert.NoError(err)

	resetEnv()
	os.Setenv(GCLOUD_KEY, gcloud_key)

	// when user sets --test, --app, --results-bucket=, and --results-dir= then
	// we should use thoe values.
	os.Setenv(GCLOUD_BUCKET, "golang-bucket")
	os.Setenv(GCLOUD_OPTIONS, "")

	app_apk_path := "/tmp/app.apk"
	if FileExists(app_apk_path) != nil {
		ioutil.WriteFile(app_apk_path, nil, 0644)
	}

	os.Setenv(APP_APK, app_apk_path)

	config, err := NewFirebaseConfig()
	config.Debug = true
	assert.NoError(err)

	gcs_object := NewGcsObjectName()
	result, err := executeGcloud(config, gcs_object)
	assert.NoError(err)

	assert.Equal([]string{
		"gcloud", "firebase", "test", "android", "run",
		"robo",
		"--app", "/tmp/app.apk",
		"--results-bucket=golang-bucket",
		"--results-dir=" + gcs_object,
	}, result)
}

func TestNewFirebaseConfig(t *testing.T) {
	assert := assert.New(t)
	resetEnv()

	err := errors.New("")

	//- APP_APK missing
	_, err = NewFirebaseConfig()
	assert.EqualError(err, "APP_APK is not defined!")

	//- APP_APK pointing to non-existent file
	os.Setenv(APP_APK, "/tmp/nope")
	_, err = NewFirebaseConfig()
	assert.EqualError(err, "file doesn't exist: '/tmp/nope'")

	//- GCLOUD_KEY missing
	os.Setenv(APP_APK, "/tmp")
	_, err = NewFirebaseConfig()
	assert.EqualError(err, "GCLOUD_KEY is not defined!")
	os.Setenv(GCLOUD_KEY, "1234")

	//- gcloud_user not defined in env or key
	_, err = NewFirebaseConfig()
	assert.EqualError(err, "GCLOUD_USER not defined in env or gcloud key")
	os.Setenv(GCLOUD_USER, "1234")

	//- gcloud_project not defined in env or key
	_, err = NewFirebaseConfig()
	assert.EqualError(err, "GCLOUD_PROJECT not defined in env or gcloud key")
	os.Setenv(GCLOUD_PROJECT, "1234")

	//- GCLOUD_BUCKET missing
	os.Setenv(GCLOUD_KEY, "1234")
	_, err = NewFirebaseConfig()
	assert.EqualError(err, "GCLOUD_BUCKET is not defined!")

	//- TEST_APK pointing to non-existent file
	os.Setenv(GCLOUD_OPTIONS, "1234")
	os.Setenv(TEST_APK, "/tmp/nope")
	_, err = NewFirebaseConfig()
	assert.EqualError(err, "file doesn't exist: '/tmp/nope'")
	os.Setenv(TEST_APK, "")

	//- HOME missing
	home := os.Getenv(HOME)
	os.Setenv(HOME, "")
	_, err = NewFirebaseConfig()
	assert.EqualError(err, "HOME is not defined!")
	os.Setenv(HOME, home)

	//- GCLOUD_KEY invalid base64
	os.Setenv(GCLOUD_KEY, " ")
	_, err = NewFirebaseConfig()
	assert.EqualError(err, "illegal base64 data at input byte 0")
	os.Setenv(GCLOUD_KEY, "1234")

	//- GCLOUD_KEY invalid path
	os.Setenv(HOME, "/does/not/exist")
	_, err = NewFirebaseConfig()
	assert.EqualError(err, "open /does/not/exist/gcloudkey.json: no such file or directory")
	os.Setenv(HOME, home)
}
