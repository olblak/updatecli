package githubaction

import (
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

var (
	// defaultWorkflowFiles specifies accepted GitHub Action workflow file name
	defaultWorkflowFiles        []string = []string{"*.yaml", "*.yml"}
	defaultVersionFilterPattern string   = "*"
	defaultVersionFilterKind    string   = "semver"
)

// Spec defines the parameters which can be provided to the Github Action crawler.
type Spec struct {
	Files []string `yaml:",omitempty"`
	// ignore allows to specify rule to ignore autodiscovery a specific Flux helmrelease based on a rule
	//
	// default: empty
	//
	Ignore MatchingRules `yaml:",omitempty"`
	// only allows to specify rule to only autodiscover manifest for a specific Flux helm release based on a rule
	//
	// default: empty
	//
	Only MatchingRules `yaml:",omitempty"`
	// OCIRepository allows to specify if OCI repository files should be updated
	//
	// default: true
	RootDir string `yaml:",omitempty"`
	// versionfilter provides parameters to specify the version pattern used when generating manifest.
	//
	// kind - semver
	//		versionfilter of kind `semver` uses semantic versioning as version filtering
	//		pattern accepts one of:
	//			`patch` - patch only update patch version
	//			`minor` - minor only update minor version
	//			`major` - major only update major versions
	//			`a version constraint` such as `>= 1.0.0`
	//
	//	kind - regex
	//		versionfilter of kind `regex` uses regular expression as version filtering
	//		pattern accepts a valid regular expression
	//
	//	example:
	//	```
	//		versionfilter:
	//			kind: semver
	//			pattern: minor
	//	```
	//
	//	and its type like regex, semver, or just latest.
	//
	VersionFilter version.Filter `yaml:",omitempty"`
	// Token allows to specify a GitHub token to use for GitHub API requests
	//
	// default:
	//   Default value is set to {{ requiredEnv "GITHUB_TOKEN }} for using the default environment variable
	Token string `yaml:",omitempty"`
}

// GitHubAction holds all information needed to generate GitHubAction manifest.
type GitHubAction struct {
	// files defines the accepted GitHub Action workflow file name
	files []string
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootDir defines the  oot directory from where looking for Flux
	rootDir string
	// scmID hold the scmID used by the newly generated manifest
	scmID string
	// versionFilter holds the "valid" version.filter, that might be different from the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
	// workflowFiles is a list of HelmRelease files found
	workflowFiles []string
	// token allows to specify a GitHub token to use for GitHub API requests
	token string
}

// New return a new valid Flux object.
func New(spec interface{}, rootDir, scmID string) (GitHubAction, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return GitHubAction{}, err
	}

	dir := rootDir
	if len(s.RootDir) > 0 {
		dir = s.RootDir
	}

	// If no RootDir have been provided via settings,
	// then fallback to the current process path.
	if len(dir) == 0 {
		logrus.Errorln("no working directory defined")
		return GitHubAction{}, err
	}

	files := defaultWorkflowFiles
	if len(s.Files) > 0 {
		files = s.Files
	}

	token := s.Token
	if token == "" {
		if os.Getenv("UPDATECLI_GITHUB_TOKEN") != "" {
			token = os.Getenv("UPDATECLI_GITHUB_TOKEN")
		} else if os.Getenv("GITHUB_TOKEN") != "" {
			token = os.Getenv("GITHUB_TOKEN")
		} else {
			logrus.Errorln("no GitHub token defined, please provide a GitHub token via the setting `token` or the UPDATECLI_GITHUB_TOKEN environment variable.")
			return GitHubAction{}, fmt.Errorf("no GitHub token defined")
		}
	}

	newFilter := s.VersionFilter
	if s.VersionFilter.IsZero() {
		// By default, helm versioning uses semantic versioning.
		newFilter.Kind = defaultVersionFilterKind
		newFilter.Pattern = defaultVersionFilterPattern
	}

	return GitHubAction{
		spec:          s,
		files:         files,
		rootDir:       dir,
		scmID:         scmID,
		versionFilter: newFilter,
		token:         token,
	}, nil

}

func (g GitHubAction) DiscoverManifests() ([][]byte, error) {
	logrus.Infof("\n\n%s\n", strings.ToTitle("GitHub Action"))
	logrus.Infof("%s\n", strings.Repeat("=", len("GitHub Action")+1))

	err := g.searchWorkflowFiles(g.rootDir, g.files)
	if err != nil {
		return nil, err
	}

	manifests := g.discoverWorkflowManifests()

	return manifests, err
}
