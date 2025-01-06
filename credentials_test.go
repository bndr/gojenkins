package gojenkins

// import (
// 	"context"
// 	"fmt"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// )

// const integration_test string = "INTEGRATION"

// var (
// 	cm         *CredentialsManager
// 	domain     = "_"
// 	dockerID   = "dockerIDCred"
// 	sshID      = "sshIdCred"
// 	usernameID = "usernameIDcred"
// 	fileID     = "fileIDcred"
// 	scope      = "GLOBAL"
// )

// func TestCreateUsernameCredentials(t *testing.T) {
// 	ctx := context.Background()
// 	j, cleanup := Setup(t, ctx)
// 	defer cleanup()

// 	cm = &CredentialsManager{J: j}
// 	cred := UsernameCredentials{
// 		ID:       usernameID,
// 		Scope:    scope,
// 		Username: "usernameTest",
// 		Password: "pass",
// 	}

// 	err := cm.Add(ctx, domain, cred)
// 	assert.Nil(t, err, "Could not create credential")

// 	getCred := UsernameCredentials{}
// 	err = cm.GetSingle(ctx, domain, cred.ID, &getCred)
// 	assert.Nil(t, err, "Could not get credential")

// 	assert.Equal(t, cred.Scope, getCred.Scope, "Scope is not equal")
// 	assert.Equal(t, cred.ID, cred.ID, "ID is not equal")
// 	assert.Equal(t, cred.Username, cred.Username, "Username is not equal")
// }

// func TestCreateFileCredentials(t *testing.T) {
// 	ctx := context.Background()
// 	j, cleanup := Setup(t, ctx)
// 	defer cleanup()

// 	cm = &CredentialsManager{J: j}
// 	cred := FileCredentials{
// 		ID:          fileID,
// 		Scope:       scope,
// 		Filename:    "testFile.json",
// 		SecretBytes: "VGhpcyBpcyBhIHRlc3Qu\n",
// 	}

// 	err := cm.Add(ctx, domain, cred)
// 	assert.Nil(t, err, "Could not create credential")

// 	getCred := FileCredentials{}
// 	err = cm.GetSingle(ctx, domain, cred.ID, &getCred)
// 	assert.Nil(t, err, "Could not get credential")

// 	assert.Equal(t, cred.Scope, getCred.Scope, "Scope is not equal")
// 	assert.Equal(t, cred.ID, cred.ID, "ID is not equal")
// 	assert.Equal(t, cred.Filename, cred.Filename, "Filename is not equal")
// }

// func TestCreateDockerCredentials(t *testing.T) {
// 	ctx := context.Background()
// 	j, cleanup := Setup(t, ctx)
// 	defer cleanup()

// 	cm = &CredentialsManager{J: j}
// 	cred := DockerServerCredentials{
// 		Scope:             scope,
// 		ID:                dockerID,
// 		Username:          "docker-name",
// 		ClientCertificate: "some secret value",
// 		ClientKey:         "client key",
// 	}

// 	err := cm.Add(ctx, domain, cred)
// 	assert.Nil(t, err, "Could not create credential")

// 	getCred := DockerServerCredentials{}
// 	err = cm.GetSingle(ctx, domain, cred.ID, &getCred)
// 	assert.Nil(t, err, "Could not get credential")

// 	assert.Equal(t, cred.Scope, getCred.Scope, "Scope is not equal")
// 	assert.Equal(t, cred.ID, cred.ID, "ID is not equal")
// 	assert.Equal(t, cred.Username, cred.Username, "Username is not equal")
// 	assert.Equal(t, cred.ClientCertificate, cred.ClientCertificate, "ClientCertificate is not equal")
// 	assert.Equal(t, cred.ClientKey, cred.ClientKey, "Username is not equal")

// }

// func TestCreateSSHCredentialsFullFlow(t *testing.T) {
// 	ctx := context.Background()
// 	j, cleanup := Setup(t, ctx)
// 	defer cleanup()

// 	cm = &CredentialsManager{J: j}
// 	sshCred := SSHCredentials{
// 		Scope:      scope,
// 		ID:         sshID,
// 		Username:   "RANDONMANE",
// 		Passphrase: "password",
// 		PrivateKeySource: &PrivateKeyFile{
// 			Value: "testValueofkey",
// 			Class: KeySourceOnMasterType,
// 		},
// 	}

// 	err := cm.Add(ctx, domain, sshCred)
// 	assert.Nil(t, err, "Could not create credential")

// 	sshCred.Username = "new_username"
// 	err = cm.Update(ctx, domain, sshCred.ID, sshCred)
// 	assert.Nil(t, err, "Could not update credential")

// 	getSSH := SSHCredentials{}
// 	err = cm.GetSingle(ctx, domain, sshCred.ID, &getSSH)
// 	assert.Nil(t, err, "Could not get ssh credential")

// 	assert.Equal(t, sshCred.Scope, getSSH.Scope, "Scope is not equal")
// 	assert.Equal(t, sshCred.ID, getSSH.ID, "ID is not equal")
// 	assert.Equal(t, sshCred.Username, getSSH.Username, "Username is not equal")
// 	assert.Equal(t, sshCred.Scope, getSSH.Scope, "Scope is not equal")

// 	err = cm.Delete(ctx, domain, getSSH.ID)
// 	assert.Nil(t, err, "Could not delete credentials")

// }

// func TestMain(t *testing.T) {
// 	//setup

// 	fmt.Printf("Debug, from TestMain\n")

// }
