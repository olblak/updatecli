package helmfile

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/condition"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	"github.com/updatecli/updatecli/pkg/plugins/resources/helm"
	"github.com/updatecli/updatecli/pkg/plugins/resources/yaml"
)

var (
	// DefaultFilePattern specifies accepted Helm chart metadata file name
	DefaultFilePattern [2]string = [2]string{"*.yaml", "*.yml"}
)

// Release specify the release information that we are looking for in Helmfile
type release struct {
	Name    string
	Chart   string
	Version string
}

// Repository specify the repository information that we are looking for in Helmfile
type repository struct {
	Name     string
	URL      string
	OCI      bool
	Username string
	Password string
}

// chartMetadata is the information that we need to retrieve from Helm chart files.
type helmfileMetadata struct {
	Name         string
	Repositories []repository
	Releases     []release
}

func (h Helmfile) discoverHelmfileReleaseManifests() ([]config.Spec, error) {

	var manifests []config.Spec

	foundHelmfileFiles, err := searchHelmfileFiles(
		h.rootDir,
		DefaultFilePattern[:])

	if err != nil {
		return nil, err
	}

	for _, foundHelmfile := range foundHelmfileFiles {

		relativeFoundChartFile, err := filepath.Rel(h.rootDir, foundHelmfile)
		if err != nil {
			// Let's try the next chart if one fail
			logrus.Errorln(err)
			continue
		}

		helmfileRelativeMetadataPath := filepath.Dir(relativeFoundChartFile)
		helmfileFilename := filepath.Base(helmfileRelativeMetadataPath)

		// Test if the ignore rule based on path is respected
		if len(h.spec.Ignore) > 0 && h.spec.Ignore.isMatchingIgnoreRule(h.rootDir, relativeFoundChartFile) {
			logrus.Debugf("Ignoring Helmfile %q from %q, as not matching rule(s)\n",
				helmfileFilename,
				helmfileRelativeMetadataPath)
			continue
		}

		// Test if the only rule based on path is respected
		if len(h.spec.Only) > 0 && !h.spec.Only.isMatchingOnlyRule(h.rootDir, relativeFoundChartFile) {
			logrus.Debugf("Ignoring Helmfile %q from %q, as not matching rule(s)\n",
				helmfileFilename,
				helmfileRelativeMetadataPath)
			continue
		}

		// Retrieve chart dependencies for each chart

		metadata, err := getHelmfileMetadata(foundHelmfile)
		if err != nil {
			return nil, err
		}

		if metadata == nil {
			continue
		}

		if len(metadata.Releases) == 0 {
			continue
		}

		for i, release := range metadata.Releases {
			manifestName := fmt.Sprintf("Bump %q Helm Chart version for Helmfile %q", release.Name, relativeFoundChartFile)

			var chartName, chartURL, OCIUsername, OCIPassword string
			var isOCI bool

			for _, repository := range metadata.Repositories {
				if strings.HasPrefix(release.Chart, repository.Name+"/") {
					chartName = strings.TrimPrefix(release.Chart, repository.Name+"/")
					chartURL = repository.URL
					isOCI = repository.OCI
					OCIUsername = repository.Username
					OCIPassword = repository.Password
					break
				}
			}

			if chartName == "" || chartURL == "" {
				logrus.Debugf("repository not identified for release %q, skipping", release.Chart)
				continue
			}

			// Helmfile uses the repository boolean "oci"
			// to identify if we are manipulating an OCI Helm chart
			// Updatecli expect the scheme "oci://".
			// Therefor we ensure that our repository URL doesn't contain http scheme
			if isOCI {
				for _, scheme := range []string{"https://", "http://"} {
					if strings.HasPrefix(chartURL, scheme) {
						chartURL = strings.TrimPrefix(chartURL, scheme)
						break
					}
				}
				chartURL = "oci://" + chartURL
			}

			if release.Version == "" {
				logrus.Debugf("no version specified for release %q, skipping", release.Chart)
				continue
			}

			helmSourcespec := helm.Spec{
				Name: chartName,
				URL:  chartURL,
			}
			if OCIUsername != "" && isOCI {
				helmSourcespec.InlineKeyChain.Username = OCIUsername
			}
			if OCIPassword != "" && isOCI {
				helmSourcespec.InlineKeyChain.Password = OCIPassword
			}

			manifest := config.Spec{
				Name: manifestName,
				Sources: map[string]source.Config{
					release.Name: {
						ResourceConfig: resource.ResourceConfig{
							Name: fmt.Sprintf("Get latest %q Helm Chart Version", release.Name),
							Kind: "helmchart",
							Spec: helmSourcespec,
						},
					},
				},
				Conditions: map[string]condition.Config{
					release.Name: {
						DisableSourceInput: true,
						ResourceConfig: resource.ResourceConfig{
							Name: fmt.Sprintf("Ensure release %q is specified for Helmfile %q", release.Name, relativeFoundChartFile),
							Kind: "yaml",
							Spec: yaml.Spec{
								File:  foundHelmfile,
								Key:   fmt.Sprintf("releases[%d].chart", i),
								Value: release.Chart,
							},
						},
					},
				},
				Targets: map[string]target.Config{
					release.Name: {
						SourceID: release.Name,
						ResourceConfig: resource.ResourceConfig{
							Name: fmt.Sprintf("Bump %q Helm Chart Version for Helmfile %q", release.Name, relativeFoundChartFile),
							Kind: "yaml",
							Spec: yaml.Spec{
								File: foundHelmfile,
								Key:  fmt.Sprintf("releases[%d].version", i),
							},
						},
					},
				},
			}
			manifests = append(manifests, manifest)

		}
	}

	return manifests, nil
}
