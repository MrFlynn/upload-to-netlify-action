package upload

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"github.com/go-openapi/runtime/client"
	"github.com/netlify/open-api/v2/go/models"
	"github.com/netlify/open-api/v2/go/plumbing/operations"
	"github.com/netlify/open-api/v2/go/porcelain"

	netlify_context "github.com/netlify/open-api/v2/go/porcelain/context"
)

// Handler provides high level functions to upload files to Netlify through their SDK.
type Handler struct {
	Token string
}

func (h Handler) createContext(inner context.Context) (outer context.Context) {
	outer = netlify_context.WithAuthInfo(inner, client.BearerToken(h.Token))
	return
}

// GetSite returns a model of the site using the given name if one exists.
func (h Handler) GetSite(ctx context.Context, name string) (site *models.Site, err error) {
	ctx = h.createContext(ctx)

	var sites []*models.Site
	sites, err = porcelain.Default.ListSites(ctx, &operations.ListSitesParams{Context: ctx, Name: &name})
	if err != nil {
		return
	}

	for _, s := range sites {
		if s.Name == name {
			site = s
			return
		}
	}

	err = fmt.Errorf("could not find site with exact name %s", name)
	return
}

// GetSiteFiles returns the list of files for a specific site.
func (h Handler) GetSiteFiles(ctx context.Context, id string) (files []*models.File, err error) {
	params := &operations.ListSiteFilesParams{
		Context: h.createContext(ctx),
		SiteID:  id,
	}

	var result *operations.ListSiteFilesOK
	result, err = porcelain.Default.Operations.ListSiteFiles(params, client.BearerToken(h.Token))
	if err != nil {
		return
	}

	files = result.GetPayload()
	return
}

// GetLatestDeploy returns the most recent deploy for the given site if one exists.
func (h Handler) GetLatestDeploy(ctx context.Context, id, branch string) (deploy *models.Deploy, err error) {
	params := &operations.ListSiteDeploysParams{
		Context: h.createContext(ctx),
		SiteID:  id,
		Branch:  &branch,
	}

	var result *operations.ListSiteDeploysOK
	result, err = porcelain.Default.Operations.ListSiteDeploys(params, client.BearerToken(h.Token))
	if err != nil {
		return
	}

	if deploys := result.GetPayload(); len(deploys) > 0 {
		deploy = deploys[0]
	} else {
		err = fmt.Errorf("could not get deploy for site_id:%s branch:%s", id, branch)
	}

	return
}

// DeployWithFilesParams contains all of the necessary parameters to initiate a new deployment.
type DeployWithFilesParams struct {
	ID     string
	Branch string
	Files  map[string]string
}

// NewDeployWithExistingFiles creates a DeployWithFilesParams object from a site ID, branch name, and a list
// of existing site files.
func NewDeployWithExistingFiles(id, branch string, assets []*models.File) (params *DeployWithFilesParams) {
	params = &DeployWithFilesParams{
		ID:     id,
		Branch: branch,
		Files:  map[string]string{},
	}

	for _, asset := range assets {
		params.Files[asset.ID] = asset.Sha
	}

	return
}

// RegisterFile adds a file path name and its hash to the file list.
func (d *DeployWithFilesParams) RegisterFile(path string, content io.ReadSeeker) (err error) {
	hash := sha1.New()

	_, err = io.Copy(hash, content)
	if err != nil {
		return
	}

	d.Files[path] = hex.EncodeToString(hash.Sum(nil))

	content.Seek(0, 0)
	return
}

// CreateDeployWithFiles creates a new site deployment.
func (h Handler) CreateDeployWithFiles(ctx context.Context, deployParams *DeployWithFilesParams) (deploy *models.Deploy, err error) {
	params := &operations.CreateSiteDeployParams{
		Context: h.createContext(ctx),
		SiteID:  deployParams.ID,
		Deploy: &models.DeployFiles{
			Branch: deployParams.Branch,
			Files:  deployParams.Files,
		},
	}

	var result *operations.CreateSiteDeployOK
	result, err = porcelain.Default.Operations.CreateSiteDeploy(params, client.BearerToken(h.Token))
	if err != nil {
		return
	}

	deploy = result.GetPayload()
	return
}

// DeployFileUploadParams contains the information necessary to upload a new file to a deploy.
type DeployFileUploadParams struct {
	DeployID string
	Path     string
	File     io.ReadCloser
}

// UploadFilesToDeploy uploads a slice of files to an open deploy on Netlify.
func (h Handler) UploadFilesToDeploy(ctx context.Context, deployFiles ...DeployFileUploadParams) (files []*models.File, err error) {
	ctx = h.createContext(ctx)
	files = make([]*models.File, 0, len(deployFiles))

	for _, deployFile := range deployFiles {
		params := &operations.UploadDeployFileParams{
			Context:  ctx,
			DeployID: deployFile.DeployID,
			Path:     deployFile.Path,
			FileBody: deployFile.File,
		}

		result, e := porcelain.Default.Operations.UploadDeployFile(params, client.BearerToken(h.Token))
		if e != nil {
			err = errors.Join(err, e)
		} else {
			files = append(files, result.GetPayload())
		}
	}

	return
}

// WaitForDeploy waits until the deploy is ready.
func (h Handler) WaitForDeploy(ctx context.Context, deploy *models.Deploy) (err error) {
	ctx = h.createContext(ctx)

	_, err = porcelain.Default.WaitUntilDeployReady(ctx, deploy)
	return
}

// DestroyDeploy cancels and then deletes the deploy with the given ID.
func (h Handler) DestroyDeploy(ctx context.Context, id string) (err error) {
	if id == "" {
		return
	}

	ctx = h.createContext(ctx)

	_, err = porcelain.Default.Operations.CancelSiteDeploy(
		&operations.CancelSiteDeployParams{
			Context:  ctx,
			DeployID: id,
		},
		client.BearerToken(h.Token),
	)

	if err != nil {
		return
	}

	_, err = porcelain.Default.Operations.DeleteDeploy(
		&operations.DeleteDeployParams{
			Context:  ctx,
			DeployID: id,
		},
		client.BearerToken(h.Token),
	)

	return
}
