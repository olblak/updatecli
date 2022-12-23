package fleet

import (
	"bytes"
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/sirupsen/logrus"
)

var (
	// FleetBundleFiles specifies accepted Helm chart metadata file name
	FleetBundleFiles [2]string = [2]string{"fleet.yaml", "fleet.yml"}
)

// Dependency specify the fleetHelmData information that we are looking for in Fleet bundle
type fleetHelmData struct {
	Chart   string
	Repo    string
	Version string
}

// fleetMetada is the information that we need to retrieve from Helm chart files.
type fleetMetada struct {
	Helm fleetHelmData
}

func (f Fleet) discoverFleetDependenciesManifests() ([][]byte, error) {

	var manifests [][]byte

	foundFleetBundleFiles, err := searchFleetBundleFiles(
		f.rootDir,
		FleetBundleFiles[:])

	if err != nil {
		return nil, err
	}

	for _, foundFleetBundleFile := range foundFleetBundleFiles {

		relativeFoundChartFile, err := filepath.Rel(f.rootDir, foundFleetBundleFile)
		if err != nil {
			// Let's try the next chart if one fail
			logrus.Errorln(err)
			continue
		}

		chartRelativeMetadataPath := filepath.Dir(relativeFoundChartFile)
		chartName := filepath.Base(chartRelativeMetadataPath)

		// Test if the ignore rule based on path is respected
		if len(f.spec.Ignore) > 0 && f.spec.Ignore.isMatchingIgnoreRule(f.rootDir, relativeFoundChartFile) {
			logrus.Debugf("Ignoring Helm Chart %q from %q, as not matching rule(s)\n",
				chartName,
				chartRelativeMetadataPath)
			continue
		}

		// Test if the only rule based on path is respected
		if len(f.spec.Only) > 0 && !f.spec.Only.isMatchingOnlyRule(f.rootDir, relativeFoundChartFile) {
			logrus.Debugf("Ignoring Helm Chart %q from %q, as not matching rule(s)\n",
				chartName,
				chartRelativeMetadataPath)
			continue
		}

		// Retrieve chart dependencies for each chart

		data, err := getFleetBundleData(foundFleetBundleFile)
		if err != nil {
			return nil, err
		}

		if data == nil {
			continue
		}

		// Skip pipeline if at least of the helm chart or helm repository is not specified
		if len(data.Helm.Chart) == 0 || len(data.Helm.Repo) == 0 {
			continue
		}

		//manifest := config.Spec{
		//	Name: manifestName,
		//	Sources: map[string]source.Config{
		//		sourceID: {
		//			ResourceConfig: resource.ResourceConfig{
		//				Name: fmt.Sprintf("Get latest %q Helm Chart Version", data.Helm.Chart),
		//				Kind: "helmchart",
		//				Spec: helm.Spec{
		//					Name: data.Helm.Chart,
		//					URL:  data.Helm.Repo,
		//				},
		//			},
		//		},
		//	},
		//	Conditions: map[string]condition.Config{
		//		conditionID + "-name": {
		//			DisableSourceInput: true,
		//			ResourceConfig: resource.ResourceConfig{
		//				Name: fmt.Sprintf("Ensure Helm chart name %q is specified", data.Helm.Chart),
		//				Kind: "yaml",
		//				Spec: yaml.Spec{
		//					File:  relativeFoundChartFile,
		//					Key:   "helm.chart",
		//					Value: data.Helm.Chart,
		//				},
		//			},
		//		},
		//		conditionID + "-repository": {
		//			DisableSourceInput: true,
		//			ResourceConfig: resource.ResourceConfig{
		//				Name: fmt.Sprintf("Ensure Helm chart repository %q is specified", data.Helm.Repo),
		//				Kind: "yaml",
		//				Spec: yaml.Spec{
		//					File:  relativeFoundChartFile,
		//					Key:   "helm.repo",
		//					Value: data.Helm.Repo,
		//				},
		//			},
		//		},
		//	},
		//	Targets: map[string]target.Config{
		//		targetID: {
		//			SourceID: sourceID,
		//			ResourceConfig: resource.ResourceConfig{
		//				Name: fmt.Sprintf("Bump chart %q from Fleet bundle %q", data.Helm.Chart, chartName),
		//				Kind: "yaml",
		//				Spec: helm.Spec{
		//					File: relativeFoundChartFile,
		//					Key:  "helm.version",
		//				},
		//			},
		//		},
		//	},
		//}

		tmpl, err := template.New("manifest").Parse(manifestTemplate)
		if err != nil {
			logrus.Errorln(err)
			continue
		}

		params := struct {
			ManifestName               string
			ImageName                  string
			ChartName                  string
			ChartRepository            string
			ConditionID                string
			FleetBundle                string
			SourceID                   string
			SourceName                 string
			SourceKind                 string
			SourceVersionFilterKind    string
			SourceVersionFilterPattern string
			TargetID                   string
			File                       string
			ScmID                      string
		}{
			ManifestName:               fmt.Sprintf("Bump Fleet Bundle %q for Helm Chart %q", chartName, data.Helm.Chart),
			ChartName:                  data.Helm.Chart,
			ChartRepository:            data.Helm.Repo,
			ConditionID:                data.Helm.Chart,
			FleetBundle:                chartName,
			SourceID:                   data.Helm.Chart,
			SourceName:                 fmt.Sprintf("Get latest %q Helm Chart Version", data.Helm.Chart),
			SourceKind:                 "helmchart",
			SourceVersionFilterKind:    "semver",
			SourceVersionFilterPattern: "*",
			TargetID:                   data.Helm.Chart,
			File:                       relativeFoundChartFile,
			ScmID:                      f.scmID,
		}

		manifest := bytes.Buffer{}
		if err := tmpl.Execute(&manifest, params); err != nil {
			return nil, err
		}

		manifests = append(manifests, manifest.Bytes())

	}

	return manifests, nil
}
