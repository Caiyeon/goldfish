package github

import (
	"errors"
	"strings"

	"github.com/google/go-github/github"
	"github.com/hashicorp/hcl"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

// if head commit is somewhere between current commit state and head of target branch,
// find all hcl files in the path folder of commit, and return contents as a map
func GetHCLFilesFromPath(accessToken, owner, repo, branch, path, base, head string) (map[string]string, error) {
	if accessToken == "" || owner == "" || repo == "" || head == "" {
		return nil, errors.New("Config_path does not include GitHub info required")
	}

	// construct oauth github client from personal access token
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	client := github.NewClient(oauth2.NewClient(ctx, ts))

	if base != "" {
		// head must be strictly ahead of base
		comparison, _, err := client.Repositories.CompareCommits(ctx, owner, repo, base, head)
		if err != nil {
			return nil, err
		}
		if comparison.GetAheadBy() < 1 || comparison.GetBehindBy() > 1 {
			return nil, errors.New("Head must be strictly ahead of base")
		}

		if branch != "" {
			// if head commit is ahead of branch, then it doesn't exist on branch and should be rejected
			comparison, _, err = client.Repositories.CompareCommits(ctx, owner, repo, head, branch)
			if err != nil {
				return nil, err
			}
			if comparison.GetBehindBy() > 1 {
				return nil, errors.New("Head must be on configured branch")
			}
		}
	}

	// grab all files in the path of the head commit
	_, folder, _, err := client.Repositories.GetContents(ctx, owner, repo, path,
		&github.RepositoryContentGetOptions{Ref: head},
	)
	if err != nil {
		return nil, err
	}
	if folder == nil || len(folder) == 0 {
		return nil, errors.New("No .hcl files found in commit and path")
	}

	// this will be the returned map if all goes well
	policies := make(map[string]string)

	// for each file, if it is a well-formed hcl file, add it to the returned array
	for _, file := range folder {

		if *file.Type == "file" && strings.HasSuffix(*file.Name, ".hcl") {
			// proposing a new root policy is not allowed. I can't believe I have to check this.
			if *file.Name == "root.hcl" {
				continue
			}

			// fetch contents of file as a string
			file, _, _, err := client.Repositories.GetContents(ctx, owner, repo, path+"/"+*file.Name,
				&github.RepositoryContentGetOptions{Ref: head},
			)
			if err != nil {
				return nil, err
			}
			content, err := file.GetContent()
			if err != nil {
				return nil, err
			}

			// verify the string is a well-formed HCL file
			if _, err := hcl.Parse(content); err != nil {
				return nil, errors.New("Could not parse " + *file.Name + " as an HCL file")
			}

			// add this file to the map to be returned
			policies[strings.TrimSuffix(*file.Name, ".hcl")] = content
		}

	}

	return policies, nil
}
