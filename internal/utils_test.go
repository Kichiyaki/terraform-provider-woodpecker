package internal_test

import (
	"testing"

	"code.gitea.io/sdk/gitea"
	"github.com/google/uuid"
	"github.com/woodpecker-ci/woodpecker/woodpecker-go/woodpecker"
)

func createRepo(tb testing.TB) *gitea.Repository {
	tb.Helper()

	repo, _, err := giteaClient.CreateRepo(gitea.CreateRepoOption{
		Name:          uuid.NewString(),
		Description:   uuid.NewString(),
		Private:       false,
		AutoInit:      true,
		Template:      false,
		License:       "MIT",
		Readme:        "Default",
		DefaultBranch: "master",
	})
	if err != nil {
		tb.Fatalf("got unexpected error while creating repo: %s", err)
	}
	tb.Cleanup(func() {
		_, _ = giteaClient.DeleteRepo(repo.Owner.UserName, repo.Name)
	})

	return repo
}

func createBranch(tb testing.TB, repo *gitea.Repository) *gitea.Branch {
	tb.Helper()

	branch, _, err := giteaClient.CreateBranch(repo.Owner.UserName, repo.Name, gitea.CreateBranchOption{
		BranchName: uuid.NewString(),
	})
	if err != nil {
		tb.Fatalf("got unexpected error while creating branch: %s", err)
	}

	return branch
}

func activateRepo(tb testing.TB, giteaRepo *gitea.Repository) *woodpecker.Repo {
	tb.Helper()

	repo, err := woodpeckerClient.RepoPost(giteaRepo.ID)
	if err != nil {
		tb.Fatalf("got unexpected error while activating repo: %s", err)
	}
	tb.Cleanup(func() {
		_ = woodpeckerClient.RepoDel(repo.ID)
	})

	return repo
}
