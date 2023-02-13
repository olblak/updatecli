package gitbranch

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source returns the latest git tag based on create time
func (gt *GitBranch) Source(workingDir string) (string, error) {

	if len(gt.spec.Path) == 0 && len(workingDir) > 0 {
		gt.spec.Path = workingDir
	}

	err := gt.Validate()
	if err != nil {
		return "", err
	}

	tags, err := gt.nativeGitHandler.Branches(workingDir)

	if err != nil {
		return "", err
	}

	gt.foundVersion, err = gt.versionFilter.Search(tags)
	if err != nil {
		return "", err
	}
	value := gt.foundVersion.GetVersion()

	if len(value) == 0 {
		logrus.Infof("%s No Git branch found matching pattern %q", result.FAILURE, gt.versionFilter.Pattern)
		return value, fmt.Errorf("no Git Branch found matching pattern %q", gt.versionFilter.Pattern)
	} else if len(value) > 0 {
		logrus.Infof("%s Git branch %q found matching pattern %q", result.SUCCESS, value, gt.versionFilter.Pattern)
	} else {
		logrus.Errorf("Something unexpected happened in gitBranch source")
	}

	return value, nil
}