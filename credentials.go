package gojenkins

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
)

// CredentialsManager is utility to control credential plugin
// Credentials declared by it can be used in jenkins jobs
type CredentialsManager struct {
	J      *Jenkins
	Folder string
}

const baseFolderPrefix = "/job/%s"
const baseCredentialsURL = "%s/credentials/store/%s/domain/%s/"
const createCredentialsURL = baseCredentialsURL + "createCredentials"
const deleteCredentialURL = baseCredentialsURL + "credential/%s/doDelete"
const configCredentialURL = baseCredentialsURL + "credential/%s/config.xml"
const credentialsListURL = baseCredentialsURL + "api/json"

var listQuery = map[string]string{
	"tree": "credentials[id]",
}

// ClassUsernameCredentials is name if java class which implements credentials that store username-password pair
const ClassUsernameCredentials = "com.cloudbees.plugins.credentials.impl.UsernamePasswordCredentialsImpl"

type credentialID struct {
	ID string `json:"id"`
}

type credentialIDs struct {
	Credentials []credentialID `json:"credentials"`
}

// UsernameCredentials struct representing credential for storing username-password pair
type UsernameCredentials struct {
	XMLName     xml.Name `xml:"com.cloudbees.plugins.credentials.impl.UsernamePasswordCredentialsImpl"`
	ID          string   `xml:"id"`
	Scope       string   `xml:"scope"`
	Description string   `xml:"description"`
	Username    string   `xml:"username"`
	Password    string   `xml:"password"`
}

// StringCredentials represents credentials that store only a secret text value.
type StringCredentials struct {
	XMLName     xml.Name `xml:"org.jenkinsci.plugins.plaincredentials.impl.StringCredentialsImpl"`
	ID          string   `xml:"id"`
	Scope       string   `xml:"scope"`
	Description string   `xml:"description"`
	Secret      string   `xml:"secret"`
}

// FileCredentials represents credentials that store a file.
// SecretBytes is a base64 encoded file content.
type FileCredentials struct {
	XMLName     xml.Name `xml:"org.jenkinsci.plugins.plaincredentials.impl.FileCredentialsImpl"`
	ID          string   `xml:"id"`
	Scope       string   `xml:"scope"`
	Description string   `xml:"description"`
	Filename    string   `xml:"fileName"`
	SecretBytes string   `xml:"secretBytes"`
}

// SSHCredentials represents credentials for SSH keys.
type SSHCredentials struct {
	XMLName          xml.Name    `xml:"com.cloudbees.jenkins.plugins.sshcredentials.impl.BasicSSHUserPrivateKey"`
	ID               string      `xml:"id"`
	Scope            string      `xml:"scope"`
	Username         string      `xml:"username"`
	Description      string      `xml:"description,omitempty"`
	PrivateKeySource interface{} `xml:"privateKeySource"`
	Passphrase       string      `xml:"passphrase,omitempty"`
}

// DockerServerCredentials represents credentials for Docker server keys.
type DockerServerCredentials struct {
	XMLName             xml.Name `xml:"org.jenkinsci.plugins.docker.commons.credentials.DockerServerCredentials"`
	ID                  string   `xml:"id"`
	Scope               string   `xml:"scope"`
	Username            string   `xml:"username"`
	Description         string   `xml:"description,omitempty"`
	ClientKey           string   `xml:"clientKey"`
	ClientCertificate   string   `xml:"clientCertificate"`
	ServerCaCertificate string   `xml:"serverCaCertificate"`
}

// KeySourceDirectEntryType is used when secret in provided directly as private key value
const KeySourceDirectEntryType = "com.cloudbees.jenkins.plugins.sshcredentials.impl.BasicSSHUserPrivateKey$DirectEntryPrivateKeySource"

// KeySourceOnMasterType is used when private key value is path to file on jenkins master
const KeySourceOnMasterType = "com.cloudbees.jenkins.plugins.sshcredentials.impl.BasicSSHUserPrivateKey$FileOnMasterPrivateKeySource"

// PrivateKey is used in SSHCredentials type.
// Class can be either KeySourceDirectEntryType (value is the secret text)
// or KeySourceOnMasterType (value is the path on master where secret is stored).
type PrivateKey struct {
	Value string `xml:"privateKey"`
	Class string `xml:"class,attr"`
}

// PrivateKeyFile represents a private key stored in a file on the Jenkins master.
type PrivateKeyFile struct {
	Value string `xml:"privateKeyFile"`
	Class string `xml:"class,attr"`
}

func (cm CredentialsManager) fillURL(url string, params ...interface{}) string {
	var args []interface{}
	if cm.Folder != "" {
		args = []interface{}{fmt.Sprintf(baseFolderPrefix, cm.Folder), "folder"}
	} else {
		args = []interface{}{"", "system"}
	}
	return fmt.Sprintf(url, append(args, params...)...)
}

// List returns the IDs of all credentials stored in the specified domain.
func (cm CredentialsManager) List(ctx context.Context, domain string) ([]string, error) {

	idsResponse := credentialIDs{}
	ids := make([]string, 0)
	err := cm.handleResponse(cm.J.Requester.Get(ctx, cm.fillURL(credentialsListURL, domain), &idsResponse, listQuery))
	if err != nil {
		return ids, err
	}

	for _, id := range idsResponse.Credentials {
		ids = append(ids, id.ID)
	}

	return ids, nil
}

// GetSingle retrieves a single credential by domain and ID.
// The credential is parsed as XML into the creds parameter (must be a pointer to struct).
func (cm CredentialsManager) GetSingle(ctx context.Context, domain string, id string, creds interface{}) error {
	str := ""
	err := cm.handleResponse(cm.J.Requester.Get(ctx, cm.fillURL(configCredentialURL, domain, id), &str, map[string]string{}))
	if err != nil {
		return err
	}

	return xml.Unmarshal([]byte(str), &creds)
}

// Add creates a new credential in the specified domain.
// The creds parameter must be a struct that can be marshaled to XML.
func (cm CredentialsManager) Add(ctx context.Context, domain string, creds interface{}) error {
	return cm.postCredsXML(ctx, cm.fillURL(createCredentialsURL, domain), creds)
}

// Delete removes a credential from the specified domain.
func (cm CredentialsManager) Delete(ctx context.Context, domain string, id string) error {
	return cm.handleResponse(cm.J.Requester.Post(ctx, cm.fillURL(deleteCredentialURL, domain, id), nil, cm.J.Raw, map[string]string{}))
}

// Update modifies an existing credential in the specified domain.
// The creds parameter must be a pointer to struct that can be marshaled to XML.
func (cm CredentialsManager) Update(ctx context.Context, domain string, id string, creds interface{}) error {
	return cm.postCredsXML(ctx, cm.fillURL(configCredentialURL, domain, id), creds)
}

func (cm CredentialsManager) postCredsXML(ctx context.Context, url string, creds interface{}) error {
	payload, err := xml.Marshal(creds)
	if err != nil {
		return err
	}

	return cm.handleResponse(cm.J.Requester.PostXML(ctx, url, string(payload), cm.J.Raw, map[string]string{}))
}

func (cm CredentialsManager) handleResponse(resp *http.Response, err error) error {
	if err != nil {
		return err
	}

	if resp.StatusCode == 409 {
		return fmt.Errorf("Resource already exists, conflict status returned")
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("invalid response code %d", resp.StatusCode)
	}

	return nil
}
