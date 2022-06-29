package release

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/drone/go-scm/scm"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/resources/gitea/client"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines settings used to interact with Gitea release
type Spec struct {
	client.Spec `yaml:",inline,omitempty"`
	// [S][C][T] Owner specifies repository owner
	Owner string `yaml:",omitempty" jsonschema:"required"`
	// [S][C][T]Repository specifies the name of a repository for a specific owner
	Repository string `yaml:",omitempty" jsonschema:"required"`
	// [S][C] VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	VersionFilter version.Filter `yaml:",omitempty"`
	// [T] Title defines the Gitea release title.
	Title string `yaml:",omitempty"`
	// [T] Tag defines the Gitea release tag.
	Tag string `yaml:",omitempty"`
	// [T] Commitish defines the commit-ish such as `main`
	Commitish string `yaml:",omitempty"`
	// [T] Draft defines if the release is a draft release
	Draft bool `yaml:",omitempty"`
	// [T] Prerelease defines if the release is a pre-release release
	Prerelease bool `yaml:",omitempty"`
}

// Gittea contains information to interact with Gitea api
type Gitea struct {
	// Spec contains inputs coming from updatecli configuration
	Spec Spec
	// client handle the api authentication
	client       client.Client
	foundVersion version.Version
	// Holds the "valid" version.filter, that might be different than the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
}

// New returns a new valid Github object.
func New(s Spec, pipelineID string) (*Gitea, error) {
	err := s.Validate()

	if err != nil {
		return &Gitea{}, err
	}

	c, err := client.New(client.Spec{
		URL:   s.URL,
		Token: s.Token,
	})

	if err != nil {
		return &Gitea{}, err
	}

	newFilter, err := s.VersionFilter.Init()
	if err != nil {
		return &Gitea{}, err
	}

	g := Gitea{
		Spec:          s,
		client:        c,
		versionFilter: newFilter,
	}

	return &g, nil
}

// Retrieve git tags from a remote gitea repository
func (g *Gitea) SearchReleases() ([]string, error) {

	ctx := context.Background()
	// Timeout api query after 30sec
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	releases, resp, err := g.client.Releases.List(
		ctx,
		strings.Join([]string{g.Spec.Owner, g.Spec.Repository}, "/"),
		scm.ReleaseListOptions{
			Page:   1,
			Size:   30,
			Open:   true,
			Closed: true,
		},
	)

	if resp.Status > 400 {
		logrus.Debugf("Gitea Api Response:\nReturn Code: %q\nBody:\n%s", resp.Status, resp.Body)
	}

	if err != nil {
		return nil, err
	}

	results := []string{}
	for i := len(releases) - 1; i >= 0; i-- {
		fmt.Printf("Title: %q\n", releases[i].Title)
		if !releases[i].Draft {
			results = append(results, releases[i].Tag)
		}
	}

	return results, nil
}

func (s *Spec) Validate() error {
	gotError := false
	missingParameters := []string{}

	err := s.ValidateClient()

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

	if (s.VersionFilter == version.Filter{}) {
		newFilter, err := s.VersionFilter.Init()
		if err != nil {
			return err
		}
		s.VersionFilter = newFilter
	}

	if len(missingParameters) > 0 {
		logrus.Errorf("missing parameter(s) [%s]", strings.Join(missingParameters, ","))
	}

	if gotError {
		return fmt.Errorf("wrong gitea configuration")
	}

	if len(s.Commitish) == 0 {
		s.Commitish = "main"
	}

	return nil
}
