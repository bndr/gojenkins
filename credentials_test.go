package gojenkins

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var (
	cm         CredentialsManager
	domain     = "_"
	dockerID   = "dockerIDCred"
	sshID      = "sshIdCred"
	usernameID = "usernameIDcred"
	scope      = "GLOBAL"
)

func TestCreateUsernameCredentials(t *testing.T) {

	cred := UsernameCredentials{
		ID:       usernameID,
		Scope:    scope,
		Username: "usernameTest",
		Password: "pass",
	}

	err := cm.Add(domain, cred)
	assert.Nil(t, err, "Could not create credential")

	getCred := UsernameCredentials{}
	err = cm.GetSingle(domain, cred.ID, &getCred)
	assert.Nil(t, err, "Could not get credential")

	assert.Equal(t, cred.Scope, getCred.Scope, "Scope is not equal")
	assert.Equal(t, cred.ID, cred.ID, "ID is not equal")
	assert.Equal(t, cred.Username, cred.Username, "Username is not equal")
}

func TestCreateDockerCredentials(t *testing.T) {

	cred := DockerServerCredentials{
		Scope:             scope,
		ID:                dockerID,
		Username:          "docker-name",
		ClientCertificate: "some secret value",
		ClientKey:         "client key",
	}

	err := cm.Add(domain, cred)
	assert.Nil(t, err, "Could not create credential")

	getCred := DockerServerCredentials{}
	err = cm.GetSingle(domain, cred.ID, &getCred)
	assert.Nil(t, err, "Could not get credential")

	assert.Equal(t, cred.Scope, getCred.Scope, "Scope is not equal")
	assert.Equal(t, cred.ID, cred.ID, "ID is not equal")
	assert.Equal(t, cred.Username, cred.Username, "Username is not equal")
	assert.Equal(t, cred.ClientCertificate, cred.ClientCertificate, "ClientCertificate is not equal")
	assert.Equal(t, cred.ClientKey, cred.ClientKey, "Username is not equal")

}

func TestCreateSSHCredentialsFullFlow(t *testing.T) {
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

	err := cm.Add(domain, sshCred)
	assert.Nil(t, err, "Could not create credential")

	sshCred.Username = "new_username"
	err = cm.Update(domain, sshCred.ID, sshCred)
	assert.Nil(t, err, "Could not update credential")

	getSSH := SSHCredentials{}
	err = cm.GetSingle(domain, sshCred.ID, &getSSH)
	assert.Nil(t, err, "Could not get ssh credential")

	assert.Equal(t, sshCred.Scope, getSSH.Scope, "Scope is not equal")
	assert.Equal(t, sshCred.ID, getSSH.ID, "ID is not equal")
	assert.Equal(t, sshCred.Username, getSSH.Username, "Username is not equal")
	assert.Equal(t, sshCred.Scope, getSSH.Scope, "Scope is not equal")

	err = cm.Delete(domain, getSSH.ID)
	assert.Nil(t, err, "Could not delete credentials")

}

func TestMain(m *testing.M) {
	//setup
	jenkins := CreateJenkins(nil, "http://localhost:8080", "admin", "admin")
	jenkins.Init()

	cm = CredentialsManager{J: jenkins}

	//execute tests
	os.Exit(m.Run())
}
