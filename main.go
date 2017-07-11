package main

import (
	"fmt"
	"os"
	// "os/exec"
	"encoding/base64"
)

func main() {
	fmt.Println("GCLOUD_USER:", os.Getenv("GCLOUD_USER")) // optional. can be read from gcloud key
	fmt.Println("GCLOUD_PROJECT:", os.Getenv("GCLOUD_PROJECT")) // optional. can be read from gcloud key
	fmt.Println("GCLOUD_KEY:", os.Getenv("GCLOUD_KEY")) // required

	// TODO: check app / test apk exist
	fmt.Println("APP_APK:", os.Getenv("APP_APK")) // required
	fmt.Println("TEST_APK:", os.Getenv("TEST_APK")) // optional. robo tests don't use a test apk

	// TODO: check gcloud exists

	gcloud_key, err := base64.StdEncoding.DecodeString(os.Getenv("GCLOUD_KEY"))
	if err != nil {
		panic(err)
	}

	// TODO: write this to $HOME/gcloudkey.json
	fmt.Println("GCLOUD_KEY decoded:", gcloud_key)

	/*
	    echo $GCLOUD_KEY | base64 --decode > "$HOME/gcloudkey.json"
        gcloud config set project "$GCLOUD_PROJECT"
        gcloud auth activate-service-account --key-file "$HOME/gcloudkey.json" "$GCLOUD_USER"

	*/

	// TODO: allow configuration options

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
