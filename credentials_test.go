package gojenkins

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const integration_test string = "INTEGRATION"

var (
	cm         *CredentialsManager
	domain     = "_"
	dockerID   = "dockerIDCred"
	sshID      = "sshIdCred"
	usernameID = "usernameIDcred"
	fileID     = "fileIDcred"
	scope      = "GLOBAL"
)

func TestCreateUsernameCredentials(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	cred := UsernameCredentials{
		ID:       usernameID,
		Scope:    scope,
		Username: "usernameTest",
		Password: "pass",
	}

	ctx := context.Background()
	err := cm.Add(ctx, domain, cred)
	assert.Nil(t, err, "Could not create credential")

	getCred := UsernameCredentials{}
	err = cm.GetSingle(ctx, domain, cred.ID, &getCred)
	assert.Nil(t, err, "Could not get credential")

	assert.Equal(t, cred.Scope, getCred.Scope, "Scope is not equal")
	assert.Equal(t, cred.ID, cred.ID, "ID is not equal")
	assert.Equal(t, cred.Username, cred.Username, "Username is not equal")
}

func TestCreateFileCredentials(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	cred := FileCredentials{
		ID:          fileID,
		Scope:       scope,
		Filename:    "testFile.json",
		SecretBytes: "VGhpcyBpcyBhIHRlc3Qu\n",
	}

	ctx := context.Background()
	err := cm.Add(ctx, domain, cred)
	assert.Nil(t, err, "Could not create credential")

	getCred := FileCredentials{}
	err = cm.GetSingle(ctx, domain, cred.ID, &getCred)
	assert.Nil(t, err, "Could not get credential")

	assert.Equal(t, cred.Scope, getCred.Scope, "Scope is not equal")
	assert.Equal(t, cred.ID, cred.ID, "ID is not equal")
	assert.Equal(t, cred.Filename, cred.Filename, "Filename is not equal")
}

func TestCreateDockerCredentials(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	cred := DockerServerCredentials{
		Scope:             scope,
		ID:                dockerID,
		Username:          "docker-name",
		ClientCertificate: "some secret value",
		ClientKey:         "client key",
	}

	ctx := context.Background()
	err := cm.Add(ctx, domain, cred)
	assert.Nil(t, err, "Could not create credential")

	getCred := DockerServerCredentials{}
	err = cm.GetSingle(ctx, domain, cred.ID, &getCred)
	assert.Nil(t, err, "Could not get credential")

	assert.Equal(t, cred.Scope, getCred.Scope, "Scope is not equal")
	assert.Equal(t, cred.ID, cred.ID, "ID is not equal")
	assert.Equal(t, cred.Username, cred.Username, "Username is not equal")
	assert.Equal(t, cred.ClientCertificate, cred.ClientCertificate, "ClientCertificate is not equal")
	assert.Equal(t, cred.ClientKey, cred.ClientKey, "Username is not equal")

}

func TestCreateSSHCredentialsFullFlow(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	sshCred := SSHCredentials{
		Scope:      scope,
		ID:         sshID,
		Username:   "RANDONMANE",
		Passphrase: "password",
		PrivateKeySource: &PrivateKeyFile{
			Value: "testValueofkey",
			Class: KeySourceOnMasterType,
		},
	}

	ctx := context.Background()
	err := cm.Add(ctx, domain, sshCred)
	assert.Nil(t, err, "Could not create credential")

	sshCred.Username = "new_username"
	err = cm.Update(ctx, domain, sshCred.ID, sshCred)
	assert.Nil(t, err, "Could not update credential")

	getSSH := SSHCredentials{}
	err = cm.GetSingle(ctx, domain, sshCred.ID, &getSSH)
	assert.Nil(t, err, "Could not get ssh credential")

	assert.Equal(t, sshCred.Scope, getSSH.Scope, "Scope is not equal")
	assert.Equal(t, sshCred.ID, getSSH.ID, "ID is not equal")
	assert.Equal(t, sshCred.Username, getSSH.Username, "Username is not equal")
	assert.Equal(t, sshCred.Scope, getSSH.Scope, "Scope is not equal")

	err = cm.Delete(ctx, domain, getSSH.ID)
	assert.Nil(t, err, "Could not delete credentials")

}

func TestMain(m *testing.M) {
	//setup
	ctx := context.Background()
	jenkins := CreateJenkins(nil, "http://localhost:8080", "admin", "admin")
	jenkins.Init(ctx)

	cm = &CredentialsManager{J: jenkins}
	fmt.Printf("Debug, from TestMain\n")
	//execute tests
	os.Exit(m.Run())
}
