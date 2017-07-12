package main

import (
	"fmt"
	"os"
	"encoding/base64"
	"io/ioutil"
	"encoding/json"
	"path"
	"github.com/bitrise-io/go-utils/command"
	"strings"
)

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func getOptionalEnv(env string) string {
	return os.Getenv(env)
}

func getRequiredEnv(env string) string {
	if len(env) == 0 {
		panic(env + " is not defined!")
	}

	return os.Getenv(env)
}

func isEmpty(str string) bool {
	return len(str) == 0
}

func checkFileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

type GcloudKeyFile struct {
	ProjectID   string `json:"project_id"`
	ClientEmail string `json:"client_email"`
}

func runCommand(cmd string) {
	cmdSlice :=  strings.Fields(cmd)

	cmdObj := command.NewWithStandardOuts(cmdSlice[0], cmdSlice[1:]...)
	checkError(cmdObj.Run())
}

func main() {
	// todo: refactor this into a method which populates a struct
	gcloud_user := ""    //getOptionalEnv("GCLOUD_USER")
	gcloud_project := "" //getOptionalEnv("GCLOUD_PROJECT")

	app_apk := getRequiredEnv("APP_APK")
	checkFileExists(app_apk)

	test_apk := getOptionalEnv("TEST_APK")
	if !isEmpty(test_apk) {
		checkFileExists(test_apk)
	}

	gcloud_key, err := base64.StdEncoding.DecodeString(getRequiredEnv("GCLOUD_KEY"))
	checkError(err)

	empty_gcloud_user := isEmpty(gcloud_user)
	empty_gcloud_project := isEmpty(gcloud_project)

	if empty_gcloud_user || empty_gcloud_project {
		parsedKeyFile := GcloudKeyFile{}
		json.Unmarshal([]byte(gcloud_key), &parsedKeyFile)

		if empty_gcloud_user {
			gcloud_user = parsedKeyFile.ClientEmail
			if isEmpty(gcloud_user) {
				panic("Missing gcloud user")
			}
		}

		if empty_gcloud_project {
			gcloud_project = parsedKeyFile.ProjectID
			if isEmpty(gcloud_project) {
				panic("Missing gcloud project")
			}
		}
	}

	fmt.Println("User: ", gcloud_user)
	fmt.Println("Project: ", gcloud_project)
	fmt.Println("App APK: ", app_apk)
	fmt.Println("Test APK: ", app_apk)

	home_dir := getRequiredEnv("HOME")
	key_file_path := path.Join(home_dir, "gcloudkey.json")
	checkError(ioutil.WriteFile(key_file_path, gcloud_key, 0644))

	runCommand("gcloud config set project " + gcloud_project)
	runCommand("gcloud auth activate-service-account --key-file " + key_file_path + " " + gcloud_user )

	// TODO: allow configuration options
	// TODO: how to grab stdout from executed command

	/*
	device_ids := "NexusLowRes"
	api_level := 25

	test_type := "robo"

	if !isEmpty(test_apk) {
		test_type = "instrumentation"
	}
	*/

	/*
    @device_ids = %w[NexusLowRes]
    @api_level  = 25

    if @robo
      type = '--type robo'
    else
      type         = '--type instrumentation'
      test_apk     = %Q(--test "#{ENV['TEST_APK']}")
      sd_card_path = '--directories-to-pull=/sdcard'
    end

    flags = [
        type,
        %Q(--app "#{ENV['APP_APK']}"),
        test_apk,
        "--results-bucket android-#{@app_name}",
        "--device-ids #{@device_ids.join(',')}",
        "--os-version-ids #{@api_level}",
        '--locales en',
        '--orientations portrait',
        '--timeout 25m',
        sd_card_path
    ].reject &:nil?

    flags << %Q(--test-targets "#{@test_targets}") unless @test_targets.empty?

    # must use custom env separator or gcloud CLI will get confused on comma separated annotations
    if @opts[:coverage] || @annotations
      env_vars      = []
      env_vars      += ['coverage=true', 'coverageFile=/sdcard/coverage.ec'] if @opts[:coverage]
      env_vars      += ["annotation=#{@annotations}"] if @annotations
      env_separator = ':'
      flags << "--environment-variables ^#{env_separator}^#{env_vars.join(env_separator)}"
    end

    gcloud firebase test android run FLAGS


    parse bucket link from gcloud CLI


    // download firebase.ec

    bucket = 'gs:/' + bucket_link.split('storage/browser').last
    bucket = "#{bucket}#{@device_ids.first}-#{@api_level}-en-portrait/artifacts/coverage.ec"
    _run_command "gsutil cp #{bucket} /bitrise/src/#{@app_name}/app/build/firebase.ec"
	*/

	// TODO: use exit code from gcloud CLI

	//
	// --- Exit codes:
	// The exit code of your Step is very important. If you return
	//  with a 0 exit code `bitrise` will register your Step as "successful".
	// Any non zero exit code will be registered as "failed" by `bitrise`.
	os.Exit(0)
}
