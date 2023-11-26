package internal_test

import (
	"sync"
	"testing"

	"code.gitea.io/sdk/gitea"
	"github.com/Kichiyaki/terraform-provider-woodpecker/internal/woodpecker"
	"github.com/google/uuid"
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

var activateRepoMu sync.Mutex

func activateRepo(tb testing.TB, giteaRepo *gitea.Repository) *woodpecker.Repo {
	tb.Helper()

	// there is a problem in Woodpecker with activating multiple repos from the same owner at the same time
	// UNIQUE constraint failed: orgs.name
	activateRepoMu.Lock()
	defer activateRepoMu.Unlock()

	repo, err := woodpeckerClient.RepoPost(giteaRepo.ID)
	if err != nil {
		tb.Fatalf("got unexpected error while activating repo: %s", err)
	}
	tb.Cleanup(func() {
		_ = woodpeckerClient.RepoDel(repo.ID)
	})

	return repo
}
