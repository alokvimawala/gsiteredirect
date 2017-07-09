# Redirect visitors to Google Site sites on Google Apps for your domain

[go-appengine]: https://cloud.google.com/appengine/docs/go/
[golang]: https://golang.org/

## Objectives

- Deploy a Slack slash command server to App Engine.
- Configure an internal Slack app to call an App Engine application for slash
  commands.

## Before you begin

1.  [Install and configure Go](https://golang.org/doc/install).
1.  Set up [App Engine and your development
    environment](https://cloud.google.com/appengine/docs/standard/go/quickstart).

## Configure the Go app

1.  Open `config.go` in a text editor.
1.  Set the token field to the value you copied. Change the line

        var oldGSiteBase string = "https://sites.google.com/a/umich.edu/"
        var newGSiteBase string = "https://sites.google.com/umich.edu/"

    to

        var oldGSiteBase string = "https://sites.google.com/a/YOURDOMAIN.SUFFIX/"
        var newGSiteBase string = "https://sites.google.com/YOURDOMAIN.SUFFIX/"

## Build and Deploy

1.  Deploy the app to App Engine.

        goapp deploy -application your-project app.yaml

    Replace `your-project` with your Google Cloud Project ID.

1.  If this is not the first App Engine application you have deployed to this
    project, go to the [Google Cloud Platform
    Console](https://console.cloud.google.com/appengine/versions), select
    version 1 in the App Engine versions and click **Migrate traffic** to send
    requests to the deployed version.
