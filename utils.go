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

func GcloudOptionsToSet(slice []string) map[string]bool {
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
func NewGcsObjectName() string {
	letterCount := 4
	bytes := make([]byte, letterCount)

	for i := 0; i < letterCount; i++ {
		bytes[i] = letters[RandomInt(len(letters))]
	}

	return time.Now().Format("2006-01-02_3:04:05.999999") + "_" + string(bytes)
}

// returns from 0 to max-1 [0, max)
func RandomInt(max int) int {
	value, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		panic(err)
	}

	return int(value.Int64())
}

func GetOptionalEnv(env string) string {
	return os.Getenv(env)
}

func GetRequiredEnv(env string) (string, error) {
	result := os.Getenv(env)
	if len(result) == 0 {
		return "", errors.New(env + " is not defined!")
	}

	return result, nil
}

func IsEmpty(str string) bool {
	return len(str) == 0
}

func FileExists(filePath string) error {
	_, err := os.Stat(filePath)
	if err != nil {
		return errors.New("file doesn't exist: '" + filePath + "'")
	}
	return nil
}

func RunCommand(cmd string) error {
	cmdSlice := strings.Fields(cmd)

	cmdObj := command.NewWithStandardOuts(cmdSlice[0], cmdSlice[1:]...)
	return cmdObj.Run()
}

func RunCommandSlice(cmdSlice []string) (int, error) {
	cmdObj := command.NewWithStandardOuts(cmdSlice[0], cmdSlice[1:]...)
	return cmdObj.RunAndReturnExitCode()
}

// Env string names
const GCLOUD_USER = "GCLOUD_USER"       // optional. read from keyfile
const GCLOUD_PROJECT = "GCLOUD_PROJECT" // optional. read from keyfile
const GCLOUD_BUCKET = "GCLOUD_BUCKET"   // required
const GCLOUD_OPTIONS = "GCLOUD_OPTIONS" // required
const APP_APK = "APP_APK"               // required
const TEST_APK = "TEST_APK"             // optional
const GCLOUD_KEY = "GCLOUD_KEY"         // required
const HOME = "HOME"

// Output from the step
const GCS_RESULTS_DIR = "GCS_RESULTS_DIR"

func FatalError(err error) {
	if err != nil {
		fmt.Println("Error: ", err.Error())
		os.Exit(1)
	}
}
