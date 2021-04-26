package pom

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/olblak/updateCli/pkg/core/scm"
	"github.com/olblak/updateCli/pkg/plugins/file"
	"github.com/sirupsen/logrus"
)

// Target reads a pom.xml and update its dependencies if needed
func (s *Spec) Target(source string, dryRun bool) (changed bool, err error) {

	data, err := file.Read(s.File, "")
	if err != nil {
		return changed, err
	}

	err = xml.Unmarshal(data, &s.project)
	if err != nil {
		return changed, err
	}

	messages, err := s.Update()
	if err != nil {
		return changed, err

	}
	if !dryRun {

		newFile, err := os.Create(filepath.Join("", s.File))
		defer newFile.Close()

		if err != nil {
			return changed, nil
		}

		output, err := xml.MarshalIndent(&s.project, "  ", "    ")
		if err != nil {
			fmt.Printf("error: %v\n", err)
		}
		_, err = newFile.Write(output)

		if err != nil {
			return changed, err
		}
	}

	if len(messages) > 0 {
		logrus.Info("\u2717 pom.xml dependencies updated")
		logrus.Info(strings.Join(messages, "\n"))
		return true, nil
	}
	logrus.Info("\u2714 pom.xml dependencies unchanged")
	return false, nil
}

// TargetFromSCM reads a pom.xml from a source control management system
// , updates the maven dependencies if needed
func (s *Spec) TargetFromSCM(source string, scm scm.Scm, dryRun bool) (changed bool, files []string, message string, err error) {
	data, err := file.Read(s.File, scm.GetDirectory())
	if err != nil {
		return changed, files, message, err
	}

	err = xml.Unmarshal(data, &s.project)
	if err != nil {
		return changed, files, message, err
	}

	messages, err := s.Update()
	if err != nil {
		return changed, files, message, err

	}

	if len(messages) == 0 {
		logrus.Info("\u2714 pom.xml dependencies unchanged")
		return changed, files, message, err
	}

	message = "[updatecli] Bump maven dependencies\n\n" + strings.Join(messages, "\n")

	logrus.Info("\u2717 pom.xml dependencies updated")
	logrus.Info(strings.Join(messages, "\n"))

	if !dryRun {

		newFile, err := os.Create(filepath.Join(scm.GetDirectory(), s.File))
		defer newFile.Close()

		if err != nil {
			return changed, files, message, nil
		}

		output, err := xml.MarshalIndent(&s.project, "  ", "    ")
		if err != nil {
			fmt.Printf("error: %v\n", err)
		}
		_, err = newFile.Write(output)

		if err != nil {
			return changed, files, message, err
		}
	}

	files = append(files, s.File)

	return changed, files, message, err
}
