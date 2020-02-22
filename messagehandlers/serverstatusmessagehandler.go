package messagehandlers

import (
	"errors"
	"fmt"
	"github.com/kpfaulkner/wheatley/models"
	"regexp"
	"strings"
)

// ServerStatusMessageHandler stores status about server (test, stage, prod etc) status
type ServerStatusMessageHandler struct {
	state models.State
}

func NewServerStatusMessageHandler() *ServerStatusMessageHandler {
	ssHandler := ServerStatusMessageHandler{}

	// load current state.
	ssHandler.state = models.LoadState()
	return &ssHandler
}

func generateEnvStateMessageString( timeStamp string, reporter string, name string, state string) string {
	response := fmt.Sprintf("At %s, %s claimed that %s was %s",  timeStamp,reporter, name, state)
	return response
}

func (ss *ServerStatusMessageHandler) doesEnvExist( env string ) bool {
	envs, err := ss.state.ListEnvs()
	if err != nil {
	  return false
	}

	lowerEnv := strings.ToLower(env)
	for _,e := range envs {
		if e == lowerEnv {
			return true
		}
	}
	return false
}

// ParseMessage takes a message, determines what to do
// return the text that should go to the user.
func (ss *ServerStatusMessageHandler) ParseMessage(msg string, user string) (string, error) {

	addStatusRegex := regexp.MustCompile(`^env (.*) is (.*)`)
	envStatusRegex := regexp.MustCompile(`^env (.*)`)
	checkStatusRegex := regexp.MustCompile(`^check env (.*)`)
	listEnvsRegex := regexp.MustCompile(`^list env`)
	summaryEnvsRegex := regexp.MustCompile(`^env summary`)
	helpRegex := regexp.MustCompile(`^help$`)
	soundOffRegex := regexp.MustCompile(`sound off`)

	switch {
	case addStatusRegex.MatchString(msg):
		res := addStatusRegex.FindStringSubmatch(msg)
		if res != nil {
			lowerEnv := strings.ToLower(res[1])
			if ss.doesEnvExist(lowerEnv) {
				ss.state.UpdateState(lowerEnv, user, res[2])
				return "if you say so", nil
			} else {
				return "don't know about the env.... " + lowerEnv, nil
			}
		}

	case summaryEnvsRegex.MatchString(msg):
		res := summaryEnvsRegex.FindStringSubmatch(msg)
		if res != nil {
			envs, err := ss.state.ListEnvs()
			if err != nil {
				return "something went boom...... sorry", nil
			}

			combinedStatus := []string{}
			for _,envName := range envs {
				env, err := ss.state.FindEnv(envName)
				if err != nil {
					// unable to find state.....  just tell the user cant do it.
					combinedStatus = append(combinedStatus, "Unable to find state for " + envName)
				}
				stateForEnv := generateEnvStateMessageString( env.Timestamp,env.Reporter, env.Name, env.State)
				combinedStatus = append( combinedStatus, stateForEnv)
			}

			return strings.Join(combinedStatus, "\n"), nil
		}

	case listEnvsRegex.MatchString(msg):
		res := listEnvsRegex.FindStringSubmatch(msg)
		if res != nil {
			envs, err := ss.state.ListEnvs()
			if err != nil {
				return "unable to list envs... something went bang...", nil
			}
			return strings.Join(envs, ","), nil
		}

	case helpRegex.MatchString(msg):
		help := []string{"list environments:  list env",
			               "set env status:       env <env name> is <whatever you want>",
			               "get env status:       env <env name>",
			               "all env summary:    env summary"}

		return strings.Join(help, "\n"), nil

	case envStatusRegex.MatchString(msg):
		res := envStatusRegex.FindStringSubmatch(msg)
		if res != nil {
      lowerEnv := strings.ToLower(res[1])
			env, err := ss.state.FindEnv(lowerEnv)
			if err != nil {
				// unable to find state.....  just tell the user cant do it.
				return "unable to find env....  ", nil
			}

			response := generateEnvStateMessageString( env.Timestamp,env.Reporter, env.Name, env.State)
			return response, nil
		}

	case checkStatusRegex.MatchString(msg):
		res := checkStatusRegex.FindStringSubmatch(msg)
		if res != nil {
			// actually goes off and performs a check itself.
			/*
			err := sg.Handler.DeleteBlock(res[2])
			if err != nil {
				return "unable to deblock", nil
			}
			*/


			return "deblocked", nil
		}

	case soundOffRegex.MatchString(msg):
		return "ServerStatusMessageHandler reporting for duty", nil

	}
	return "", errors.New("No match")

}
