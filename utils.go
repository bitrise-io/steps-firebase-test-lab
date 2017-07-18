package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/bitrise-io/go-utils/command"
	"math/big"
	"os"
	"strings"
	"time"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func gcloudOptionsToSet(slice []string) map[string]bool {
	set := make(map[string]bool)

	for i := range slice {
		// --results-bucket=/a/b/c => --results-bucket=
		key := strings.SplitAfter(slice[i], "=")[0]
		set[key] = true
	}

	return set
}

// Matches api_lib/firebase/test/arg_validate.py _GenerateUniqueGcsObjectName from gcloud SDK
// Example output: 2017-07-12_11:36:12.467586_XVlB
func newGcsObjectName() string {
	letterCount := 4
	bytes := make([]byte, letterCount)

	for i := 0; i < letterCount; i++ {
		bytes[i] = letters[randomInt(len(letters))]
	}

	return time.Now().Format("2006-01-02_3:04:05.999999") + "_" + string(bytes)
}

// randomInt returns from 0 to max-1 [0, max)
func randomInt(max int) int {
	value, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		panic(err)
	}

	return int(value.Int64())
}

func getOptionalEnv(env string) string {
	return os.Getenv(env)
}

func getRequiredEnv(env string) (string, error) {
	result := os.Getenv(env)
	if len(result) == 0 {
		return "", errors.New(env + " is not defined!")
	}

	return result, nil
}

func isEmpty(str string) bool {
	return len(str) == 0
}

func fileExists(filePath string) error {
	_, err := os.Stat(filePath)
	if err != nil {
		return errors.New("file doesn't exist: '" + filePath + "'")
	}
	return nil
}

func runCommand(cmd string) error {
	cmdSlice := strings.Fields(cmd)

	cmdObj := command.NewWithStandardOuts(cmdSlice[0], cmdSlice[1:]...)
	return cmdObj.Run()
}

func runCommandSlice(cmdSlice []string) (int, error) {
	cmdObj := command.NewWithStandardOuts(cmdSlice[0], cmdSlice[1:]...)
	return cmdObj.RunAndReturnExitCode()
}

// Env string names

const envKeyGcloudUser = "GCLOUD_USER"       // optional. read from keyfile
const envKeyGcloudProject = "GCLOUD_PROJECT" // optional. read from keyfile
const envKeyGcloudBucket = "GCLOUD_BUCKET"   // required
const envKeyGcloudOptions = "GCLOUD_OPTIONS" // required
const envKeyAppApk = "APP_APK"               // required
const envKeyTestApk = "TEST_APK"             // optional
const envKeyGcloud = "GCLOUD_KEY"            // required
const envKeyHome = "HOME"

func fatalError(err error) {
	if err != nil {
		fmt.Println("Error: ", err.Error())
		os.Exit(1)
	}
}
