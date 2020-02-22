package models

import (
    "encoding/json"
    "errors"
    "fmt"
    "io/ioutil"
    "os"
    "time"
)

var stateFileName = "state.json"

type Environment struct {
    Reporter  string `json:"reporter"`  // who reported it.
    State     string `json:"state"`   // state. general free text... keep it clean :P
    Timestamp string `json:"timestamp"`
    Name string `json:"name"`  // env name
}

type State struct {
    Environments []*Environment `json:"Environments"`
}

func LoadState() State {
    var state State
    stateFile, err := os.Open(stateFileName)
    defer stateFile.Close()
    if err != nil {
        fmt.Println(err.Error())
    }
    jsonParser := json.NewDecoder(stateFile)
    jsonParser.Decode(&state)
    return state
}


func (s *State) Save() error {
    b, err := json.Marshal(s)
    if err != nil {
        return err
    }

    err = ioutil.WriteFile(stateFileName, b, 0777)
    // handle this error
    if err != nil {
        return err
    }
    return nil
}

func (s *State) ListEnvs() ([]string, error) {
    envs := []string{}
    for _,e := range s.Environments {
        envs = append(envs, e.Name)
    }
    return envs, nil
}


// linear lookup. Yuck... but very limited number of envs.
// not concerned.
func (s *State) FindEnv(env string) (*Environment, error) {
    for _,e := range s.Environments {
        if e.Name == env {
            return e, nil
        }
    }

    // no env.
    return nil, errors.New("Unknown env")
}

// linear lookup. Yuck... but very limited number of envs.
// not concerned.
func (s *State) UpdateState(envName string, who string, state string)  error {

    env, err := s.FindEnv(envName)
    if err != nil {
        return nil
    }

    env.State = state
    env.Reporter = who

    // hack.... windows timezone issue and binaries.
    env.Timestamp = time.Now().UTC().Add( time.Hour*-7).Format("2006-01-02T15:04:05")
    // save the sucker!
    s.Save()
    return nil
}



