package maven

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

var (
	// ErrRepositoryNotFound contains the err return no maven repository found
	ErrRepositoryNotFound = errors.New("repository not found")
)

type Repository struct {
	UniqueVersion bool             `xml:"uniqueVersion"`
	Releases      RepositoryPolicy `xml:"releases"`
	Snapshots     RepositoryPolicy `xml:"snapshots"`
	ID            string           `xml:"id"`
	Name          string           `xml:"name"`
	URL           string           `xml:"url"`
	Layout        string           `xml:"layout"`
}

// https://repo1.maven.org/maven2/

// Versions returns every versions for a specific artifact from a repository
func (r *Repository) Versions(URL, groupID, artifactID string) (string, []string, error) {

	// versions contains the list of available version
	type versions struct {
		ID      xml.Name `xml:"versions"`
		Version []string `xml:"version"`
	}
	// version hold version information
	type version struct {
		Versioning xml.Name `xml:"versioning"`
		Latest     string   `xml:"latest"`
		Release    string   `xml:"release"`
		Versions   versions `xml:"versions"`
	}

	// metadata hold maven repository metadata
	type metadata struct {
		Metadata   xml.Name `xml:"metadata"`
		GroupID    string   `xml:"groupId"`
		ArtifactID string   `xml:"artifactId"`
		Versioning version  `xml:"versioning"`
	}

	URL = fmt.Sprintf("https://%s/%s/%s/maven-metadata.xml",
		URL,
		strings.ReplaceAll(groupID, ".", "/"),
		artifactID)

	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return "", nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", nil, err
	}

	if res.StatusCode > 399 {
		switch res.StatusCode {
		case 404:
			logrus.Errorf("URL: %s", URL)
			logrus.Error(ErrRepositoryNotFound.Error())
			return "", nil, err
		default:
			err = errors.New("wrong http return code")
			logrus.Errorf("URL: %s",
				URL)
			logrus.Errorf(err.Error()+": %s",
				res.StatusCode)
			return "", nil, err

		}
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", nil, err
	}

	fmt.Println(string(body))

	data := metadata{}

	err = xml.Unmarshal(body, &data)
	if err != nil {
		return "", nil, err
	}

	return data.Versioning.Latest, data.Versioning.Versions.Version, nil

}
