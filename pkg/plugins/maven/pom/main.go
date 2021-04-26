package pom

import (
	"fmt"
	"strings"

	"github.com/olblak/updateCli/pkg/plugins/maven"
)

// Spec describe parameters to manipulate to manipulate a pom.xml from updatecli
type Spec struct {
	ModelVersion string   // ModelVersion specifies the pom version currently set to 4
	File         string   // File defines the pom.xml file path
	Ignores      []string // Ignores contains a list of dependencies that we don't want to update
	Batch        int      // Batch defines the number of dependencies update per pipeline run
	project      maven.Project
}

// Validate validates a maven pom spec
func (s *Spec) Validate() error {
	if len(s.File) == 0 {
		s.File = "pom.xml"
	}

	if len(s.ModelVersion) == 0 {
		s.ModelVersion = "4"
	}
	return nil
}

// Update updates a project dependencies
func (s *Spec) Update() ([]string, error) {

	messages := []string{}

	counter := 0
	for _, dep := range s.project.Dependencies {
		if counter > s.Batch {
			break
		}

		oldVersion := dep.Version

		dep.Update(s.project.Repositories)

		newVersion := dep.Version

		if strings.Compare(oldVersion, newVersion) != 0 {
			messages = append(
				messages, fmt.Sprintf(
					"%q updated from %q to %q",
					dep.GroupID,
					oldVersion,
					newVersion))
		}
	}

	return messages, nil
}
