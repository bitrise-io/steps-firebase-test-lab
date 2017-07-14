package main

import (
	"errors"
	. "github.com/bitrise-community/steps-firebase-test-lab/utils"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
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
	Setenv(HOME, home)
	Setenv(PATH, path)
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

	Setenv(KEY, ENV_VALUE)
	value, err := GetRequiredEnv(KEY)
	assert.Equal(ENV_VALUE, value)
}

func TestExecuteGcloud(t *testing.T) {
	assert := assert.New(t)
	gcloud_key, err := GetRequiredEnv(GCLOUD_KEY)
	assert.NoError(err)

	resetEnv()
	Setenv(GCLOUD_KEY, gcloud_key)

	Setenv(GCLOUD_BUCKET, "golang-bucket")
	Setenv(GCLOUD_OPTIONS, `--device-ids NexusLowRes
	 --os-version-ids 25
	 --locales en
	 --orientations portrait
	 --timeout 25m
	 --directories-to-pull=/sdcard
	 --environment-variables ^:^coverage=true:coverageFile=/sdcard/coverage.ec`)

	app_apk_path := "/tmp/app.apk"
	test_apk_path := "/tmp/test.apk"

	if FileExists(app_apk_path) != nil {
		err = ioutil.WriteFile(app_apk_path, nil, 0644)
		PanicOnErr(err)
	}

	if FileExists(test_apk_path) != nil {
		err = ioutil.WriteFile(test_apk_path, nil, 0644)
		PanicOnErr(err)
	}

	Setenv(APP_APK, app_apk_path)
	Setenv(TEST_APK, test_apk_path)

	config, err := NewFirebaseConfig()
	config.Debug = true
	assert.NoError(err)

	gcs_object := NewGcsObjectName()
	result, err := buildGcloudCommand(config, gcs_object)
	assert.NoError(err)

	assert.Equal([]string{
		"gcloud", "firebase", "test", "android", "run",
		"--type", "instrumentation",
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
	Setenv(GCLOUD_KEY, gcloud_key)

	// when user sets --test, --app, --results-bucket=, and --results-dir= then
	// we should use thoe values.
	Setenv(GCLOUD_BUCKET, "golang-bucket")
	Setenv(GCLOUD_OPTIONS, `
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
		err = ioutil.WriteFile(app_apk_path, nil, 0644)
		PanicOnErr(err)
	}

	if FileExists(test_apk_path) != nil {
		err = ioutil.WriteFile(test_apk_path, nil, 0644)
		PanicOnErr(err)
	}

	Setenv(APP_APK, app_apk_path)
	Setenv(TEST_APK, test_apk_path)

	config, err := NewFirebaseConfig()
	config.Debug = true
	assert.NoError(err)

	gcs_object := NewGcsObjectName()
	result, err := buildGcloudCommand(config, gcs_object)
	assert.NoError(err)

	assert.Equal([]string{
		"gcloud", "firebase", "test", "android", "run",
		"--type", "instrumentation",
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
	Setenv(GCLOUD_KEY, gcloud_key)

	// when user sets --test, --app, --results-bucket=, and --results-dir= then
	// we should use thoe values.
	Setenv(GCLOUD_BUCKET, "golang-bucket")
	Setenv(GCLOUD_OPTIONS, "")

	app_apk_path := "/tmp/app.apk"
	if FileExists(app_apk_path) != nil {
		err := ioutil.WriteFile(app_apk_path, nil, 0644)
		PanicOnErr(err)
	}

	Setenv(APP_APK, app_apk_path)

	config, err := NewFirebaseConfig()
	config.Debug = true
	assert.NoError(err)

	gcs_object := NewGcsObjectName()
	result, err := buildGcloudCommand(config, gcs_object)
	assert.NoError(err)

	assert.Equal([]string{
		"gcloud", "firebase", "test", "android", "run",
		"--type", "robo",
		"--app", "/tmp/app.apk",
		"--results-bucket=golang-bucket",
		"--results-dir=" + gcs_object,
	}, result)
}

func PanicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

// handle err to pass checkerr check
func Setenv(key, value string) {
	err := os.Setenv(key, value)
	PanicOnErr(err)
}

func TestNewFirebaseConfig(t *testing.T) {
	assert := assert.New(t)
	resetEnv()

	err := errors.New("")

	//- APP_APK missing
	_, err = NewFirebaseConfig()
	assert.EqualError(err, "APP_APK is not defined!")

	//- APP_APK pointing to non-existent file
	Setenv(APP_APK, "/tmp/nope")
	_, err = NewFirebaseConfig()
	assert.EqualError(err, "file doesn't exist: '/tmp/nope'")

	//- GCLOUD_KEY missing
	Setenv(APP_APK, "/tmp")
	_, err = NewFirebaseConfig()
	assert.EqualError(err, "GCLOUD_KEY is not defined!")
	Setenv(GCLOUD_KEY, "1234")

	//- gcloud_user not defined in env or key
	_, err = NewFirebaseConfig()
	assert.EqualError(err, "GCLOUD_USER not defined in env or gcloud key")
	Setenv(GCLOUD_USER, "1234")

	//- gcloud_project not defined in env or key
	_, err = NewFirebaseConfig()
	assert.EqualError(err, "GCLOUD_PROJECT not defined in env or gcloud key")
	Setenv(GCLOUD_PROJECT, "1234")

	//- GCLOUD_BUCKET missing
	Setenv(GCLOUD_KEY, "1234")
	_, err = NewFirebaseConfig()
	assert.EqualError(err, "GCLOUD_BUCKET is not defined!")

	//- TEST_APK pointing to non-existent file
	Setenv(GCLOUD_OPTIONS, "1234")
	Setenv(TEST_APK, "/tmp/nope")
	_, err = NewFirebaseConfig()
	assert.EqualError(err, "file doesn't exist: '/tmp/nope'")
	Setenv(TEST_APK, "")

	//- HOME missing
	home := os.Getenv(HOME)
	Setenv(HOME, "")
	_, err = NewFirebaseConfig()
	assert.EqualError(err, "HOME is not defined!")
	Setenv(HOME, home)

	//- GCLOUD_KEY invalid base64
	Setenv(GCLOUD_KEY, " ")
	_, err = NewFirebaseConfig()
	assert.EqualError(err, "illegal base64 data at input byte 0")
	Setenv(GCLOUD_KEY, "1234")

	//- GCLOUD_KEY invalid path
	Setenv(HOME, "/does/not/exist")
	_, err = NewFirebaseConfig()
	assert.EqualError(err, "open /does/not/exist/gcloudkey.json: no such file or directory")
	Setenv(HOME, home)
}
