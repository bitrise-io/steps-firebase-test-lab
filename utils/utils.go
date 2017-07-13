package utils

import (
	"os"
	"errors"
	"github.com/bitrise-io/go-utils/command"
	"strings"
	"time"
	"math/rand"
	"fmt"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Matches api_lib/firebase/test/arg_validate.py _GenerateUniqueGcsObjectName from gcloud SDK
// Example output: 2017-07-12_11:36:12.467586_XVlB
func NewGcsObjectName() string {
	letterCount := 4
	bytes := make([]byte, letterCount)

	for i := 0; i < letterCount; i++ {
		bytes[i] = letters[rand.Intn(len(letters))]
	}

	return time.Now().Format("2006-01-02_3:04:05.999999") + "_" + string(bytes)
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
