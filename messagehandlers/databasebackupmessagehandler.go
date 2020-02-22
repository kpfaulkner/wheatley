package messagehandlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kpfaulkner/wheatley/helper"
	"os"
	"regexp"
	"strings"
	"time"
)


type DBConfig struct {
	SubscriptionID string   `json:"SubscriptionID"`
	TenantID       string   `json:"TenantID"`
	ClientID       string   `json:"ClientID"`
	ClientSecret   string   `json:"ClientSecret"`
	ResourceGroup  string   `json:"ResourceGroup"`
	StorageKey string       `json:"StorageKey"`
	StorageURL string       `json:"StorageURL"`
	SqlAdminLogin string    `json:"SqlAdminLogin"`
	SqlAdminPassword string `json:"SqlAdminPassword"`
	AllowedUsers string     `json:"AllowedUsers"`
  BackupPrefix string     `json:"BackupPrefix"`
	ServerName string       `json:"ServerName"`
	DatabaseName string     `json:"DatabaseName"`

	AllowedUsersList []string

}

type DatabaseBackupMessageHandler struct {
	asHelper *helper.AzureSQLHelper

	// config specific to test LPC.
  config *DBConfig
}

func NewDatabaseBackupMessageHandler() *DatabaseBackupMessageHandler {
	asHandler := DatabaseBackupMessageHandler{}
	asHandler.config,_ = loadDBConfig("azuredb.json")
	asHandler.asHelper = helper.NewAzureSQLHelper(asHandler.config.SubscriptionID, asHandler.config.TenantID, asHandler.config.ClientID,
																								asHandler.config.ClientSecret, asHandler.config.SqlAdminLogin, asHandler.config.SqlAdminPassword,
																								asHandler.config.StorageKey, asHandler.config.StorageURL, asHandler.config.ResourceGroup )
  return &asHandler
}

func loadDBConfig(filename string ) (*DBConfig, error)  {
	configFile, err := os.Open(filename)
	defer configFile.Close()
	if err != nil {
		return nil, err
	}

	config := DBConfig{}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)

	// split users for later checking.
	config.AllowedUsersList = strings.Split(config.AllowedUsers,",")
  return &config, nil
}

func userAllowed(user string, allowedUsers []string) bool {
	lowerUser := strings.ToLower(user)
  for _,au := range allowedUsers {
  	if strings.ToLower(au) == lowerUser {
  		return true
	  }
  }

  return false
}

// ParseMessage takes a message, determines what to do
// return the text that should go to the user.
func (ss *DatabaseBackupMessageHandler) ParseMessage(msg string, user string) (string, error) {

	// cant be arsed extracting out term.

	backupProdRegex := regexp.MustCompile(`^backup prod`)
	soundOffRegex := regexp.MustCompile(`sound off`)
	helpRegex := regexp.MustCompile(`^help$`)

	msg = strings.ToLower(msg)
	switch {

	case backupProdRegex.MatchString(msg):
		res := backupProdRegex.FindStringSubmatch(msg)
		if res != nil  {

			if !userAllowed(user, ss.config.AllowedUsersList) {
				return "Sorry not permitted to do this.", nil
			}

			backupName := fmt.Sprintf("%s-%s.bacpac", ss.config.BackupPrefix, time.Now().Format("2006-01-02"))
			err := ss.asHelper.StartDBExport(ss.config.ServerName, ss.config.DatabaseName, backupName)
			if err != nil {
				return "Cannot backup database!!\n", nil
			}

			return "Have started backup. There is no indication of when it will complete though.", nil
		}

	case soundOffRegex.MatchString(msg):
		return "DatabaseBackupMessageHandler reporting for duty", nil

	case helpRegex.MatchString(msg):
		return "backup prod: Starts backing up production database to blob storage.", nil

	}
	return "", errors.New("No match")

}
