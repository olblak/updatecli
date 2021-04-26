package maven

import (
	"testing"
)

// https://repo1.maven.org/maven2/foxtrot/foxtrot-core/

type DataSet []Data
type Data struct {
	dep             Dependency
	expectedVersion string
	expectedError   error
	repos           []Repository
}

var (
	dataSet = DataSet{
		{
			dep: Dependency{
				GroupID:    "org.jenkins-ci.main",
				ArtifactID: "jenkins-war",
			},
			expectedVersion: "2.288",

			repos: []Repository{
				{
					URL: "repo.jenkins-ci.org/releases",
				},
			},
		},
		{
			dep: Dependency{
				GroupID:    "foxtrot",
				ArtifactID: "foxtrot-core",
			},
			expectedVersion: "4.0",
		},
		{
			dep: Dependency{
				GroupID:    "foxtrot",
				ArtifactID: "foxtrot-core",
			},
			expectedVersion: "4.0",

			repos: []Repository{
				{
					URL: "repo1.maven.org/maven2",
				},
			},
		},
	}
)

func TestUpdate(t *testing.T) {

	for _, data := range dataSet {
		err := data.dep.Update(data.repos)
		if err != nil {
			t.Errorf("unexpected error: %q", err.Error())
		}
		if data.dep.Version != data.expectedVersion {
			t.Errorf("Version:\n\tExpected:\t%q\n\tGo:\t\t%q", data.expectedVersion, data.dep.Version)
		}

	}
}
