package main

import (
	"testing"
	"os"
	"io/ioutil"
)

// os.Exit(1) = test passes.

// GCLOUD_USER    // Optional. Read from keyfile
// GCLOUD_PROJECT // Optional. Read from keyfile
// GCLOUD_BUCKET  // Required
// GCLOUD_OPTIONS // Required
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

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func TestHello(t *testing.T) {
	gcloud_key, err := getRequiredEnv(GCLOUD_KEY)
	panicOnError(err)

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

	test_apk_path := "/tmp/app.apk"
	app_apk_path := "/tmp/test.apk"

	if fileExists(test_apk_path) != nil {
		ioutil.WriteFile(test_apk_path, nil, 0644)
	}

	if fileExists(app_apk_path) != nil {
		ioutil.WriteFile(app_apk_path, nil, 0644)
	}

	os.Setenv(APP_APK, test_apk_path)
	os.Setenv(TEST_APK, app_apk_path)

	err = executeGcloud(true)
	if err != nil { t.Error(err) }
}

// t.Error("")
