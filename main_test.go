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

// envKeyGcloudUser    // Optional. Read from keyfile
// envKeyGcloudProject // Optional. Read from keyfile
// envKeyGcloudBucket  // Required
// envKeyGcloudOptions // Optional
// envKeyAppApk        // Required
// envKeyTestApk       // Optional
// envKeyGcloud     // Required

const Path = "PATH"

// envman complains if the .envstore doesn't exist however we don't want to check it into git
func init() {
	WriteFile(".envstore.yml")
	gcloudKeyValue := base64.StdEncoding.EncodeToString([]byte(`{"project_id": "fake-project","client_email": "fake@example.com"}`))

	Setenv(envKeyGcloud, gcloudKeyValue)
}

func resetEnv() {
	homeValue := os.Getenv(envKeyHome)
	pathValue := os.Getenv(Path)

	os.Clearenv()
	Setenv(envKeyHome, homeValue)
	Setenv(Path, pathValue)
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
	gcloudKeyValue, err := getRequiredEnv(envKeyGcloud)
	assert.NoError(err)

	resetEnv()
	Setenv(envKeyGcloud, gcloudKeyValue)

	Setenv(envKeyGcloudBucket, "golang-bucket")
	Setenv(envKeyGcloudOptions, `--device-ids NexusLowRes
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

	Setenv(envKeyAppApk, appApkPath)
	Setenv(envKeyTestApk, testApkPath)

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
	gcloudKeyValue, err := getRequiredEnv(envKeyGcloud)
	assert.NoError(err)

	resetEnv()
	Setenv(envKeyGcloud, gcloudKeyValue)

	// when user sets --test, --app, --results-bucket=, and --results-dir= then
	// we should use those values.
	Setenv(envKeyGcloudBucket, "golang-bucket")
	Setenv(envKeyGcloudOptions, `
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

	Setenv(envKeyAppApk, appApkPath)
	Setenv(envKeyTestApk, testApkPath)

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
	gcloudKeyValue, err := getRequiredEnv(envKeyGcloud)
	assert.NoError(err)

	resetEnv()
	Setenv(envKeyGcloud, gcloudKeyValue)

	// when user sets --test, --app, --results-bucket=, and --results-dir= then
	// we should use thoe values.
	Setenv(envKeyGcloudBucket, "golang-bucket")
	Setenv(envKeyGcloudOptions, "")

	appApkPath := "/tmp/app.apk"
	WriteFile(appApkPath)

	Setenv(envKeyAppApk, appApkPath)

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

	//- envKeyAppApk missing
	_, err = newFirebaseConfig()
	assert.EqualError(err, envKeyAppApk+" is not defined!")

	//- envKeyAppApk pointing to non-existent file
	Setenv(envKeyAppApk, "/tmp/nope")
	_, err = newFirebaseConfig()
	assert.EqualError(err, "file doesn't exist: '/tmp/nope'")

	//- envKeyGcloud missing
	Setenv(envKeyAppApk, "/tmp")
	_, err = newFirebaseConfig()
	assert.EqualError(err, envKeyGcloud+" is not defined!")
	Setenv(envKeyGcloud, "e30K") // {} base64 encoded

	//- gcloud_user not defined in env or key
	_, err = newFirebaseConfig()
	assert.EqualError(err, envKeyGcloudUser+" not defined in env or gcloud key")
	Setenv(envKeyGcloudUser, "1234")

	//- gcloud_project not defined in env or key
	_, err = newFirebaseConfig()
	assert.EqualError(err, envKeyGcloudProject+" not defined in env or gcloud key")
	Setenv(envKeyGcloudProject, "1234")

	//- envKeyGcloudBucket missing
	Setenv(envKeyGcloud, "1234")
	_, err = newFirebaseConfig()
	assert.EqualError(err, envKeyGcloudBucket+" is not defined!")

	//- envKeyTestApk pointing to non-existent file
	Setenv(envKeyGcloudOptions, "1234")
	Setenv(envKeyTestApk, "/tmp/nope")
	_, err = newFirebaseConfig()
	assert.EqualError(err, "file doesn't exist: '/tmp/nope'")
	Setenv(envKeyTestApk, "")

	//- envKeyHome missing
	homeValue := os.Getenv(envKeyHome)
	Setenv(envKeyHome, "")
	_, err = newFirebaseConfig()
	assert.EqualError(err, envKeyHome+" is not defined!")
	Setenv(envKeyHome, homeValue)

	//- envKeyGcloud invalid base64
	Setenv(envKeyGcloud, " ")
	_, err = newFirebaseConfig()
	assert.EqualError(err, "illegal base64 data at input byte 0")
	Setenv(envKeyGcloud, "1234")

	//- envKeyGcloud invalid path
	Setenv(envKeyHome, "/does/not/exist")
	_, err = newFirebaseConfig()
	assert.EqualError(err, "open /does/not/exist/gcloudkey.json: no such file or directory")
	Setenv(envKeyHome, homeValue)
}
