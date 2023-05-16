package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	"github.com/mrflynn/upload-to-netlify-action/internal/actions"
	"github.com/mrflynn/upload-to-netlify-action/internal/upload"
)

// Program information set during compile.
var version, commit, date string

// Program logger.
var logger = actions.NewLogger()

// Program variables.
var (
	siteName string

	sourceFiles      []string
	destinationPaths []string
)

// Netlify handler
var handler upload.Handler

func init() {
	logger.Debug(fmt.Sprintf(
		"upload-to-netlify-action %s (commit: %s, compiled: %s)",
		version, commit, date,
	))

	var (
		netlifyToken string
		err          error

		opts = actions.GetInputOptions{
			Required:       true,
			TrimWhitespace: true,
		}
	)

	netlifyToken, err = actions.GetInput("netlify-token", opts)
	if err != nil {
		logger.Error("Netlify API token is required!")
		os.Exit(1)
	} else {
		logger.SetSecret(netlifyToken)
	}

	siteName, err = actions.GetInput("site-name", opts)
	if err != nil {
		logger.Error("Name of Netlify site is required!")
		os.Exit(1)
	}

	sourceFiles, err = actions.GetMultilineInput("source-file", opts)
	if err != nil {
		logger.Error("At least one source file must be given.")
		os.Exit(1)
	}

	destinationPaths, err = actions.GetMultilineInput("destination-path", opts)
	if err != nil {
		logger.Error("At least one destination path must be given.")
		os.Exit(1)
	}

	handler = upload.Handler{Token: netlifyToken}
}

func printError(err error) {
	logger.Error(err.Error())
	os.Exit(1)
}

func cleanDestinationPath(path string) (cleaned string) {
	if regexp.MustCompile("[#?]").MatchString(path) {
		printError(
			fmt.Errorf("path %s contains one of the following illegal characters: #, ?", path),
		)
	}

	cleaned = strings.TrimPrefix(path, "/")
	return
}

func getReadersForSourceFiles() (rs map[string]io.ReadSeekCloser) {
	rs = make(map[string]io.ReadSeekCloser, len(destinationPaths))

	for i, sourceFile := range sourceFiles {
		file, err := os.Open(sourceFile)
		if err != nil {
			printError(err)
		}

		rs[cleanDestinationPath(destinationPaths[i])] = file
	}

	return
}

const branchName = "master" // TODO: determine best way to get this value.

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Get site information.
	site, err := handler.GetSite(ctx, siteName)
	if err != nil {
		printError(err)
	}

	logger.Debug(fmt.Sprintf("Got site ID for %s (ID: %s)", siteName, site.ID))

	// Get latest deploy and wait until it has completed.
	deploy, err := handler.GetLatestDeploy(ctx, site.ID, branchName)
	if err != nil {
		printError(err)
	}

	logger.Debug(fmt.Sprintf("Got latest deploy for site ID %s (deployID: %s)", site.ID, deploy.ID))

	err = handler.WaitForDeploy(ctx, deploy)
	if err != nil {
		printError(err)
	}

	// Get site files.
	files, err := handler.GetSiteFiles(ctx, site.ID)
	if err != nil {
		printError(err)
	}

	logger.Debug(fmt.Sprintf("Got %d preexisting files from site ID %s", len(files), site.ID))

	var (
		sourceFileReaders = getReadersForSourceFiles()
		deployParams      = upload.NewDeployWithExistingFiles(site.ID, branchName, files)
	)

	for path, reader := range sourceFileReaders {
		err = deployParams.RegisterFile("/"+path, reader)
		if err != nil {
			printError(err)
		}

		logger.Debug(fmt.Sprintf("Registered file %s", path))
	}

	logger.Info("Beginning upload of the following files: " + strings.Join(sourceFiles, ", ") + ".")

	// Create new deploy with additional files.
	deploy, err = handler.CreateDeployWithFiles(ctx, deployParams)
	if err != nil {
		printError(err)
	}

	logger.Debug(fmt.Sprintf("Started new deploy with ID %s", deploy.ID))

	uploadParams := make([]upload.DeployFileUploadParams, 0, len(destinationPaths))

	for path, reader := range sourceFileReaders {
		uploadParams = append(uploadParams, upload.DeployFileUploadParams{
			DeployID: deploy.ID,
			Path:     path,
			File:     reader,
		})
	}

	// Upload additional files and wait for deploy to finish.
	files, err = handler.UploadFilesToDeploy(ctx, uploadParams...)
	if err != nil {
		printError(err)
	}

	logger.Debug(fmt.Sprintf("Uploaded %d files to deploy with ID %s", len(files), deploy.ID))

	err = handler.WaitForDeploy(ctx, deploy)
	if err != nil {
		printError(err)
	}

	logger.Info("Files successfully uploaded to Netlify!")
}
