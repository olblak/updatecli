package branch

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/drone/go-scm/scm"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/resources/gitea/client"
	"github.com/updatecli/updatecli/pkg/plugins/utils/redact"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines settings used to interact with Gitea release
type Spec struct {
	client.Spec `yaml:",inline,omitempty"`
	// [S][C] Owner specifies repository owner
	Owner string `yaml:",omitempty" jsonschema:"required"`
	// [S][C] Repository specifies the name of a repository for a specific owner
	Repository string `yaml:",omitempty" jsonschema:"required"`
	// [S] VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	VersionFilter version.Filter `yaml:",omitempty"`
	// [C] Branch specifies the branch name
	Branch string `yaml:",omitempty"`
}

// Gitea contains information to interact with Gitea api
type Gitea struct {
	// spec contains inputs coming from updatecli configuration
	spec Spec
	// client handle the api authentication
	client        client.Client
	HeadBranch    string
	foundVersion  version.Version
	versionFilter version.Filter
}

// New returns a new valid Gitea object.
func New(spec interface{}) (*Gitea, error) {

	var s Spec
	var clientSpec client.Spec

	// mapstructure.Decode cannot handle embedded fields
	// hence we decode it in two steps
	err := mapstructure.Decode(spec, &clientSpec)
	if err != nil {
		return &Gitea{}, err
	}

	err = mapstructure.Decode(spec, &s)
	if err != nil {
		return &Gitea{}, nil
	}

	err = clientSpec.Validate()
	if err != nil {
		return &Gitea{}, err
	}

	err = clientSpec.Sanitize()
	if err != nil {
		return &Gitea{}, err
	}

	s.Spec = clientSpec

	err = s.Validate()

	if err != nil {
		return &Gitea{}, err
	}

	c, err := client.New(clientSpec)

	if err != nil {
		return &Gitea{}, err
	}

	newFilter, err := s.VersionFilter.Init()
	if err != nil {
		return &Gitea{}, err
	}
	s.VersionFilter = newFilter

	g := Gitea{
		spec:          s,
		client:        c,
		versionFilter: newFilter,
	}

	return &g, nil

}

// Retrieve gitea branches from a remote gitea repository
func (g *Gitea) SearchBranches() (tags []string, err error) {

	// Timeout api query after 30sec
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	results := []string{}
	page := 0
	for {
		branches, resp, err := g.client.Git.ListBranches(
			ctx,
			strings.Join([]string{g.spec.Owner, g.spec.Repository}, "/"),
			scm.ListOptions{
				URL:  g.spec.URL,
				Page: page,
				Size: 30,
			},
		)

		if err != nil {
			return nil, err
		}

		if resp.Status > 400 {
			logrus.Debugf("RC: %q\nBody:\n%s", resp.Status, resp.Body)
		}

		for _, branch := range branches {
			results = append(results, branch.Name)
		}

		if page >= resp.Page.Last {
			break
		}
		page++
	}

	return results, nil
}

func (s Spec) Validate() error {
	gotError := false
	missingParameters := []string{}

	err := s.Spec.Validate()

	if err != nil {
		gotError = true
	}

	if len(s.Owner) == 0 {
		gotError = true
		missingParameters = append(missingParameters, "owner")
	}

	if len(s.Repository) == 0 {
		gotError = true
		missingParameters = append(missingParameters, "repository")
	}

	if len(missingParameters) > 0 {
		logrus.Errorf("missing parameter(s) [%s]", strings.Join(missingParameters, ","))
	}

	if gotError {
		return fmt.Errorf("wrong gitea configuration")
	}

	return nil
}

// ReportConfig returns a new configuration object with only the necessary fields
// to identify the resource without any sensitive information or context specific data.
func (g *Gitea) ReportConfig() interface{} {
	return Spec{
		Owner:         g.spec.Owner,
		Repository:    g.spec.Repository,
		VersionFilter: g.spec.VersionFilter,
		Branch:        g.spec.Branch,
		Spec: client.Spec{
			URL: redact.URL(g.spec.URL),
		},
	}
}
