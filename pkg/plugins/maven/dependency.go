package maven

import (
	"strings"

	"github.com/sirupsen/logrus"
)

type Dependency struct {
	GroupID    string      `xml:"groupId"`
	ArtifactID string      `xml:"artifactId"`
	Version    string      `xml:"version"`
	Type       string      `xml:"type"`
	Classifier string      `xml:"classifier"`
	Scope      string      `xml:"scope"`
	SystemPath string      `xml:"systemPath"`
	Exclusions []Exclusion `xml:"exclusions>exclusion"`
	Optional   string      `xml:"optional"`
}

// Update fetchs the latest version of a dependency from a maven repository
func (d *Dependency) Update(repositories []Repository) error {

	mavenCentral := Repository{
		URL: "repo1.maven.org/maven2",
	}
	mavenCentralUsed := false
	// Test if mavenCentral is specified
	for _, repo := range repositories {
		if strings.Compare(repo.URL, mavenCentral.URL) == 0 {
			mavenCentralUsed = true
			break
		}
	}

	if !mavenCentralUsed {
		repositories = append(repositories, mavenCentral)
	}

	for _, repo := range repositories {
		latest, _, err := repo.Versions(
			repo.URL,
			d.GroupID,
			d.ArtifactID)

		if err == ErrRepositoryNotFound {
			continue
		} else if err != nil {
			return err
		}

		if len(latest) != 0 {
			oldVersion := d.Version
			d.Version = latest
			logrus.Debugf("Version updated from %q to %q", oldVersion, latest)
			break

		}

	}

	return nil
}
