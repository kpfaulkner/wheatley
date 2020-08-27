package helper

import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"strings"
)



type GithubHelper struct {
	client *github.Client
  ctx context.Context
	owner string
	token string
}

func NewGithubHelper(owner string, apikey string) *GithubHelper {
	gh := GithubHelper{}
	gh.owner = owner
	gh.token = apikey
  gh.ctx = context.Background()
	gh.client = getClient(gh.token)
	return &gh
}

func getClient(token string) *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
  return client
}

func (gh *GithubHelper) GetBranchesForRepo(repo string) ( []*github.Branch, error) {

	branches, _, err := gh.client.Repositories.ListBranches(gh.ctx, gh.owner, repo, nil)
	if err != nil {
		fmt.Printf("error %s\n", err.Error())
		return nil, err
	}
	return branches, nil
}


func (gh *GithubHelper) GetMergeCommentsBetweenCommits(repo string, commit1 string, commit2 string) ( []string, error) {

	cc, _, err := gh.client.Repositories.CompareCommits(gh.ctx, gh.owner, repo, commit1, commit2 )
	if err != nil {
		fmt.Printf("error %s\n", err.Error())
		return nil, err
	}

	commentSlice := []string{}
	for _, commit := range cc.Commits {
		commentSlice = append(commentSlice, *commit.GetCommit().Message)
	}

	return commentSlice,nil
}

func (gh *GithubHelper) GetPR(repo string, prID int) ( *github.PullRequest, error) {
	pr, _, err := gh.client.PullRequests.Get(gh.ctx, gh.owner, repo, prID)
	if err != nil {
		fmt.Printf("error %s\n", err.Error())
		return nil, err
	}
	return pr, nil
}

func (gh *GithubHelper) GetIssueState(repo string, issueID int) (string, error) {

	issue, err := gh.GetIssue(repo, issueID)
	if err != nil {
		return "", err
	}

	return issue.GetState(), nil
}

func (gh *GithubHelper) GetIssue(repo string, issueID int) (*github.Issue, error) {
	issue, _, err := gh.client.Issues.Get(gh.ctx, gh.owner, repo, issueID)
	if err != nil {
		fmt.Printf("error %s\n", err.Error())
		return nil, err
	}
	return issue, nil
}


func (gh *GithubHelper) addLabelToIssue(repo string,id int, label string) error {

	_,_,err :=gh.client.Issues.AddLabelsToIssue(gh.ctx, gh.owner, repo,id, []string{label})
	if err != nil {
		fmt.Printf("addLabelToIssue error %s\n", err.Error())
		return err
	}
	return nil
}

func doesLabelExist( label string, labelList []string) bool {
	for _,l := range labelList {
		if strings.ToLower(l) == label {
			// already attached.  return
			return true
		}
	}
  return false
}

func (gh *GithubHelper) AddLabelToPR(repo string, prID int, label string) error {
  pr, err := gh.GetPR(repo, prID)
  if err != nil {
  	fmt.Printf("error %s\n", err.Error())
  	return err
  }

  allLabels := make([]string,len( pr.Labels), len(pr.Labels))
  for _,l := range pr.Labels {
  	allLabels = append(allLabels, strings.ToLower(*l.Name))
  }

	lowerLabel := strings.ToLower(label)
  if !doesLabelExist(lowerLabel, allLabels) {
		gh.addLabelToIssue(repo, prID, label)
  }

  return nil
}

func (gh *GithubHelper) AddLabelToIssue(repo string, issueID int, label string) error {
	issue, err := gh.GetIssue(repo, issueID)
	if err != nil {
		fmt.Printf("error %s\n", err.Error())
		return err
	}

	state := issue.GetState()
	fmt.Printf("state is %s\n", state)
	allLabels := make([]string,len( issue.Labels), len(issue.Labels))
	for _,l := range issue.Labels {
		allLabels = append(allLabels, strings.ToLower(*l.Name))
	}

	lowerLabel := strings.ToLower(label)
	if !doesLabelExist(lowerLabel, allLabels) {
		gh.addLabelToIssue(repo, issueID, label)
	}

  return nil
}

func (gh *GithubHelper) AddUserToIssue(repo string, issueID int, user string) error {

	gh.client.Issues.AddAssignees(gh.ctx, gh.owner, repo, issueID, []string{user})
	return nil
}






