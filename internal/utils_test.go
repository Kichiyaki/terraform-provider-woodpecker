package internal_test

import (
	"strconv"
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

func createOrgRepo(tb testing.TB) *gitea.Repository {
	tb.Helper()

	org, _, err := giteaClient.CreateOrg(gitea.CreateOrgOption{
		Name:                      uuid.NewString(),
		FullName:                  uuid.NewString(),
		Visibility:                gitea.VisibleTypePublic,
		RepoAdminChangeTeamAccess: true,
	})
	if err != nil {
		tb.Fatalf("got unexpected error while creating org: %s", err)
	}
	tb.Cleanup(func() {
		_, _ = giteaClient.DeleteOrg(org.UserName)
	})

	repo, _, err := giteaClient.CreateOrgRepo(org.UserName, gitea.CreateRepoOption{
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

func createOrg(tb testing.TB) *woodpecker.Org {
	tb.Helper()

	repo := createOrgRepo(tb)
	activateRepo(tb, repo)

	org, err := woodpeckerClient.OrgLookup(repo.Owner.UserName)
	if err != nil {
		tb.Fatalf("got unexpected error while looking up org: %s", err)
	}

	return org
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

	repo, err := woodpeckerClient.RepoPost(woodpecker.RepoPostOptions{
		ForgeRemoteID: strconv.FormatInt(giteaRepo.ID, 10),
	})
	if err != nil {
		tb.Fatalf("got unexpected error while activating repo: %s", err)
	}
	tb.Cleanup(func() {
		_ = woodpeckerClient.RepoDel(repo.ID)
	})

	return repo
}
