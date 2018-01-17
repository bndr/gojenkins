package gojenkins

import (
	"encoding/xml"
	"fmt"
	"net/http"
)

//CredentialsManager is utility to control credential plugin
//Credentials declared by it can be used in jenkins jobs
type CredentialsManager struct {
	J *Jenkins
}

const baseCredentialsURL = "/credentials/store/system/domain/%s/"
const createCredentialsURL = baseCredentialsURL + "createCredentials"
const deleteCredentialURL = baseCredentialsURL + "credential/%s/delete"
const configCredentialURL = baseCredentialsURL + "credential/%s/config.xml"
const credentialsListURL = baseCredentialsURL + "api/json"

//ClassUsernameCredentials is name if java class which implements credentials that store username-password pair
const ClassUsernameCredentials = "com.cloudbees.plugins.credentials.impl.UsernamePasswordCredentialsImpl"

type credentialID struct {
	ID string `json:"id"`
}

type credentialIDs struct {
	Credentials []credentialID `json:"credentials"`
}

//UsernameCredentials struct representing credential for storing username-password pair
type UsernameCredentials struct {
	XMLName  xml.Name `xml:"com.cloudbees.plugins.credentials.impl.UsernamePasswordCredentialsImpl"`
	ID       string   `xml:"id"`
	Scope    string   `xml:"scope"`
	Username string   `xml:"username"`
	Password string   `xml:"password"`
}

var listQuery = map[string]string{
	"tree": "credentials[id]",
}

//List ids if credentials stored inside provided domain
func (cm CredentialsManager) List(domain string) ([]string, error) {

	idsResponse := credentialIDs{}
	ids := make([]string, 0)
	err := cm.handleResponse(cm.J.Requester.Get(fmt.Sprintf(credentialsListURL, domain), &idsResponse, listQuery))
	if err != nil {
		return ids, err
	}

	for _, id := range idsResponse.Credentials {
		ids = append(ids, id.ID)
	}

	return ids, nil
}

//GetSingle searches for credential in given domain with given id, if credential is found
//it will be parsed as xml to creds parameter(creds must be pointer to struct)
func (cm CredentialsManager) GetSingle(domain string, id string, creds interface{}) error {

	return cm.handleResponse(cm.J.Requester.Get(fmt.Sprintf(configCredentialURL, domain, id), creds, map[string]string{}))
}

//Add credential to given domain, creds must be struct which is parsable to xml
func (cm CredentialsManager) Add(domain string, creds interface{}) error {

	return cm.postCredsXML(fmt.Sprintf(createCredentialsURL, domain), creds)
}

//Delete credential in given domain with given id
func (cm CredentialsManager) Delete(domain string, id string) error {
	return cm.handleResponse(cm.J.Requester.PostXML(fmt.Sprintf(deleteCredentialURL, domain, id), "", cm.J.Raw, map[string]string{}))
}

//Update credential in given domain with given id, creds must be pointer to struct which is parsable to xml
func (cm CredentialsManager) Update(domain string, id string, creds interface{}) error {

	return cm.postCredsXML(fmt.Sprintf(configCredentialURL, domain, id), creds)
}

func (cm CredentialsManager) postCredsXML(url string, creds interface{}) error {
	payload, err := xml.Marshal(creds)
	if err != nil {
		return err
	}

	return cm.handleResponse(cm.J.Requester.PostXML(url, string(payload), cm.J.Raw, map[string]string{}))
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
