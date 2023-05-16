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
	logger.Debugf(
		"upload-to-netlify-action %s (commit: %s, compiled: %s)",
		version, commit, date,
	)

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

func handleError(ctx context.Context, err error, deployID *string) {
	logger.Error(err.Error())

	if deployID != nil {
		if destroyErr := handler.DestroyDeploy(ctx, *deployID); destroyErr != nil {
			logger.Errorf("Error while trying to destroy deploy: %s", destroyErr)
		}
	}

	os.Exit(1)
}

func cleanDestinationPath(path string) (cleaned string, err error) {
	if regexp.MustCompile("[#?]").MatchString(path) {
		err = fmt.Errorf("path %s contains one of the following illegal characters: #, ?", path)
		return
	}

	cleaned = strings.TrimPrefix(path, "/")
	return
}

func getReadersForSourceFiles() (rs map[string]io.ReadSeekCloser, err error) {
	rs = make(map[string]io.ReadSeekCloser, len(destinationPaths))

	var (
		file *os.File
		dest string
	)

	for i, sourceFile := range sourceFiles {
		file, err = os.Open(sourceFile)
		if err != nil {
			return
		}

		dest, err = cleanDestinationPath(destinationPaths[i])
		if err != nil {
			return
		}

		rs[dest] = file
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
		handleError(ctx, err, nil)
	}

	logger.Debugf("Got site ID for %s (ID: %s)", siteName, site.ID)

	// Get latest deploy and wait until it has completed.
	deploy, err := handler.GetLatestDeploy(ctx, site.ID, branchName)
	if err != nil {
		handleError(ctx, err, nil)
	}

	logger.Debugf("Got latest deploy for site ID %s (deployID: %s)", site.ID, deploy.ID)

	err = handler.WaitForDeploy(ctx, deploy)
	if err != nil {
		handleError(ctx, err, nil)
	}

	// Get site files.
	files, err := handler.GetSiteFiles(ctx, site.ID)
	if err != nil {
		handleError(ctx, err, nil)
	}

	logger.Debugf("Got %d preexisting files from site ID %s", len(files), site.ID)

	sourceFileReaders, err := getReadersForSourceFiles()
	if err != nil {
		handleError(ctx, err, nil)
	}

	deployParams := upload.NewDeployWithExistingFiles(site.ID, branchName, files)
	for path, reader := range sourceFileReaders {
		err = deployParams.RegisterFile("/"+path, reader)
		if err != nil {
			handleError(ctx, err, nil)
		}

		logger.Debugf("Registered file %s", path)
	}

	logger.Infof(
		"Beginning upload of the following files: %s.", strings.Join(sourceFiles, ", "),
	)

	// Create new deploy with additional files.
	deploy, err = handler.CreateDeployWithFiles(ctx, deployParams)
	if err != nil {
		handleError(ctx, err, &deploy.ID)
	}

	logger.Debugf("Started new deploy with ID %s", deploy.ID)

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
		handleError(ctx, err, &deploy.ID)
	}

	logger.Debugf("Uploaded %d files to deploy with ID %s", len(files), deploy.ID)

	err = handler.WaitForDeploy(ctx, deploy)
	if err != nil {
		handleError(ctx, err, &deploy.ID)
	}

	logger.Info("Files successfully uploaded to Netlify!")
}
