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
	repo string
	owner string
	token string
}

func NewGithubHelper(owner string, repo string, apikey string) *GithubHelper {
	gh := GithubHelper{}
	gh.owner = owner
	gh.repo = repo
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

func (gh *GithubHelper) GetPR(prID int) ( *github.PullRequest, error) {
	pr, _, err := gh.client.PullRequests.Get(gh.ctx, gh.owner, gh.repo, prID)
	if err != nil {
		fmt.Printf("error %s\n", err.Error())
		return nil, err
	}
	return pr, nil
}

func (gh *GithubHelper) GetIssueState(issueID int) (string, error) {

	issue, err := gh.GetIssue(issueID)
	if err != nil {
		return "", err
	}

	return issue.GetState(), nil
}

func (gh *GithubHelper) GetIssue(issueID int) (*github.Issue, error) {
	issue, _, err := gh.client.Issues.Get(gh.ctx, gh.owner, gh.repo, issueID)
	if err != nil {
		fmt.Printf("error %s\n", err.Error())
		return nil, err
	}
	return issue, nil
}


func (gh *GithubHelper) addLabelToIssue(id int, label string) error {

	_,_,err :=gh.client.Issues.AddLabelsToIssue(gh.ctx, gh.owner, gh.repo,id, []string{label})
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

func (gh *GithubHelper) AddLabelToPR(prID int, label string) error {
  pr, err := gh.GetPR(prID)
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
		gh.addLabelToIssue(prID, label)
  }

  return nil
}

func (gh *GithubHelper) AddLabelToIssue(issueID int, label string) error {
	issue, err := gh.GetIssue(issueID)
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
		gh.addLabelToIssue(issueID, label)
	}

  return nil
}

func (gh *GithubHelper) AddUserToIssue(issueID int, user string) error {

	gh.client.Issues.AddAssignees(gh.ctx, gh.owner, gh.repo, issueID, []string{user})
	return nil
}






