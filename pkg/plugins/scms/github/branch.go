package github

import (
	"context"
	"fmt"

	"github.com/shurcooL/githubv4"
)

// BranchesQuery defines the structure for the GraphQL query to list branches
type BranchesQuery struct {
	Repository struct {
		Refs struct {
			Nodes []struct {
				Name string
			}
			PageInfo struct {
				HasNextPage bool
				EndCursor   githubv4.String
			}
		} `graphql:"refs(refPrefix: \"refs/heads/\", first: $first, after: $after)"`
	} `graphql:"repository(owner: $owner, name: $name)"`
}

// ListBranches lists all branches from a specific repository using pagination
func ListBranches(client GitHubClient, owner, repo string, ctx context.Context) ([]string, error) {
	var branches []string
	var after *githubv4.String

	for {
		var q BranchesQuery
		vars := map[string]interface{}{
			"owner": githubv4.String(owner),
			"name":  githubv4.String(repo),
			"first": githubv4.Int(100),
			"after": after,
		}

		err := client.Query(ctx, &q, vars)
		if err != nil {
			return nil, fmt.Errorf("failed to list branches: %w", err)
		}

		for _, node := range q.Repository.Refs.Nodes {
			branches = append(branches, node.Name)
		}

		if q.Repository.Refs.PageInfo.HasNextPage {
			after = &q.Repository.Refs.PageInfo.EndCursor
		} else {
			break
		}
	}

	return branches, nil
}
