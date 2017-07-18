package main

import (
	"encoding/base64"
	"errors"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

// os.Exit(1) = test passes.

// gcloudUser    // Optional. Read from keyfile
// gcloudProject // Optional. Read from keyfile
// gcloudBucket  // Required
// gcloudOptions // Optional
// appApk        // Required
// testApk       // Optional
// gcloudKey     // Required

const Path = "Path"

// envman complains if the .envstore doesn't exist however we don't want to check it into git
func init() {
	WriteFile(".envstore.yml")
	gcloudKey := base64.StdEncoding.EncodeToString([]byte(`{"project_id": "fake-project","client_email": "fake@example.com"}`))

	Setenv(gcloudKey, gcloudKey)
}

func resetEnv() {
	home := os.Getenv(home)
	path := os.Getenv(Path)

	os.Clearenv()
	Setenv(home, home)
	Setenv(Path, path)
}

func TestFileExists(t *testing.T) {
	assert := assert.New(t)

	err := fileExists("/tmp/nope.txt")
	assert.EqualError(err, "file doesn't exist: '/tmp/nope.txt'")

	err = fileExists("/tmp")
	assert.Equal(nil, err)
}

func TestRunCommand(t *testing.T) {
	assert := assert.New(t)

	err := runCommand("true")
	assert.Equal(nil, err)
}

func TestGetRequiredEnv(t *testing.T) {
	assert := assert.New(t)

	const Key = "KEY_THAT_IS_NOT_USED"
	const EnvValue = "EnvValue"
	_, err := getRequiredEnv(Key)
	assert.EqualError(err, Key+" is not defined!")

	Setenv(Key, EnvValue)
	value, err := getRequiredEnv(Key)
	assert.Equal(EnvValue, value)
}

func TestExecuteGcloud(t *testing.T) {
	assert := assert.New(t)
	gcloudKey, err := getRequiredEnv(gcloudKey)
	assert.NoError(err)

	resetEnv()
	Setenv(gcloudKey, gcloudKey)

	Setenv(gcloudBucket, "golang-bucket")
	Setenv(gcloudOptions, `--device-ids NexusLowRes
	 --os-version-ids 25
	 --locales en
	 --orientations portrait
	 --timeout 25m
	 --directories-to-pull=/sdcard
	 --environment-variables ^:^coverage=true:coverageFile=/sdcard/coverage.ec`)

	appApkPath := "/tmp/app.apk"
	testApkPath := "/tmp/test.apk"

	WriteFile(appApkPath)
	WriteFile(testApkPath)

	Setenv(appApk, appApkPath)
	Setenv(testApk, testApkPath)

	config, err := newFirebaseConfig()
	config.Debug = true
	assert.NoError(err)

	gcsObject := newGcsObjectName()
	result, err := buildGcloudCommand(config, gcsObject)
	assert.NoError(err)

	assert.Equal([]string{
		"gcloud", "firebase", "test", "android", "run",
		"--type", "instrumentation",
		"--test", "/tmp/test.apk",
		"--app", "/tmp/app.apk",
		"--results-bucket=golang-bucket",
		"--results-dir=" + gcsObject,
		"--device-ids", "NexusLowRes",
		"--os-version-ids", "25",
		"--locales", "en",
		"--orientations", "portrait",
		"--timeout", "25m",
		"--directories-to-pull=/sdcard",
		"--environment-variables", "^:^coverage=true:coverageFile=/sdcard/coverage.ec",
	}, result)
}

func WriteFile(filePath string) {
	if fileExists(filePath) != nil {
		err := ioutil.WriteFile(filePath, nil, 0644)
		PanicOnErr(err)
	}
}

func TestExecuteGcloudUserOverrides(t *testing.T) {
	assert := assert.New(t)
	gcloudKey, err := getRequiredEnv(gcloudKey)
	assert.NoError(err)

	resetEnv()
	Setenv(gcloudKey, gcloudKey)

	// when user sets --test, --app, --results-bucket=, and --results-dir= then
	// we should use those values.
	Setenv(gcloudBucket, "golang-bucket")
	Setenv(gcloudOptions, `
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

	appApkPath := "/tmp/app.apk"
	testApkPath := "/tmp/test.apk"

	WriteFile(appApkPath)
	WriteFile(testApkPath)

	Setenv(appApk, appApkPath)
	Setenv(testApk, testApkPath)

	config, err := newFirebaseConfig()
	config.Debug = true
	assert.NoError(err)

	gcsObject := newGcsObjectName()
	result, err := buildGcloudCommand(config, gcsObject)
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
	gcloudKey, err := getRequiredEnv(gcloudKey)
	assert.NoError(err)

	resetEnv()
	Setenv(gcloudKey, gcloudKey)

	// when user sets --test, --app, --results-bucket=, and --results-dir= then
	// we should use thoe values.
	Setenv(gcloudBucket, "golang-bucket")
	Setenv(gcloudOptions, "")

	appApkPath := "/tmp/app.apk"
	WriteFile(appApkPath)

	Setenv(appApk, appApkPath)

	config, err := newFirebaseConfig()
	config.Debug = true
	assert.NoError(err)

	gcsObject := newGcsObjectName()
	result, err := buildGcloudCommand(config, gcsObject)
	assert.NoError(err)

	assert.Equal([]string{
		"gcloud", "firebase", "test", "android", "run",
		"--type", "robo",
		"--app", "/tmp/app.apk",
		"--results-bucket=golang-bucket",
		"--results-dir=" + gcsObject,
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

	//- appApk missing
	_, err = newFirebaseConfig()
	assert.EqualError(err, "appApk is not defined!")

	//- appApk pointing to non-existent file
	Setenv(appApk, "/tmp/nope")
	_, err = newFirebaseConfig()
	assert.EqualError(err, "file doesn't exist: '/tmp/nope'")

	//- gcloudKey missing
	Setenv(appApk, "/tmp")
	_, err = newFirebaseConfig()
	assert.EqualError(err, "gcloudKey is not defined!")
	Setenv(gcloudKey, "e30K") // {} base64 encoded

	//- gcloud_user not defined in env or key
	_, err = newFirebaseConfig()
	assert.EqualError(err, "gcloudUser not defined in env or gcloud key")
	Setenv(gcloudUser, "1234")

	//- gcloud_project not defined in env or key
	_, err = newFirebaseConfig()
	assert.EqualError(err, "gcloudProject not defined in env or gcloud key")
	Setenv(gcloudProject, "1234")

	//- gcloudBucket missing
	Setenv(gcloudKey, "1234")
	_, err = newFirebaseConfig()
	assert.EqualError(err, "gcloudBucket is not defined!")

	//- testApk pointing to non-existent file
	Setenv(gcloudOptions, "1234")
	Setenv(testApk, "/tmp/nope")
	_, err = newFirebaseConfig()
	assert.EqualError(err, "file doesn't exist: '/tmp/nope'")
	Setenv(testApk, "")

	//- home missing
	home := os.Getenv(home)
	Setenv(home, "")
	_, err = newFirebaseConfig()
	assert.EqualError(err, "home is not defined!")
	Setenv(home, home)

	//- gcloudKey invalid base64
	Setenv(gcloudKey, " ")
	_, err = newFirebaseConfig()
	assert.EqualError(err, "illegal base64 data at input byte 0")
	Setenv(gcloudKey, "1234")

	//- gcloudKey invalid path
	Setenv(home, "/does/not/exist")
	_, err = newFirebaseConfig()
	assert.EqualError(err, "open /does/not/exist/gcloudkey.json: no such file or directory")
	Setenv(home, home)
}
