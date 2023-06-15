package af_client

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/casdoor/casdoor/util"
)

type PtAF struct {
	url   string
	Token string

	httpClient *http.Client
}

func NewPtAF(url string) *PtAF {
	result := &PtAF{url: url}

	//dev only. should be removed in production
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	result.httpClient = &http.Client{Transport: tr}

	return result
}

func (af PtAF) doRequest(request http.Request) (http.Response, error) {
	request.Header.Set("Authorization", "Bearer "+af.Token)
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")

	resp, err := af.httpClient.Do(&request)
	return *resp, err
}

func (af PtAF) Login(request LoginRequest) (*LoginResponse, error) {
	//dev only. should be remove in production
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	body := strings.NewReader(util.StructToJson(request))
	req, _ := http.NewRequest("POST", af.url+"auth/refresh_tokens", body)

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	responseContent := string(response)

	result := &LoginResponse{}

	err = util.JsonToStruct(responseContent, result)
	if err != nil {
		return nil, err
	}

	af.Token = result.AccessToken
	return result, nil
}

func (af PtAF) GetTenant(request GetTenantRequest) (*Tenant, error) {
	body := strings.NewReader(util.StructToJson(request))
	req, _ := http.NewRequest("GET", af.url+"auth/tenants", body)

	resp, err := af.doRequest(*req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	responseContent := string(response)

	result := &Tenant{}

	err = util.JsonToStruct(responseContent, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (af PtAF) CreateTenant(request CreateTenantRequest) (*Tenant, error) {

	body := strings.NewReader(util.StructToJson(request))
	req, _ := http.NewRequest("POST", af.url+"auth/tenants", body)

	resp, err := af.doRequest(*req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	responseContent := string(response)

	result := &Tenant{}

	err = util.JsonToStruct(responseContent, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (af PtAF) CreateUser(request CreateUserRequest) error {

	body := strings.NewReader(util.StructToJson(request))
	req, err := http.NewRequest("POST", af.url+"auth/users", body)
	if err != nil {
		return fmt.Errorf("http.NewRequest: %w", err)
	}

	resp, err := af.doRequest(*req)
	if err != nil {
		return fmt.Errorf("af.doRequest: %w", err)
	}

	response, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("ioutil.ReadAll: %w", err)
	}

	responseContent := string(response)

	if resp.StatusCode != 201 {
		return fmt.Errorf("af.doRequest response status code: %d and body: %s", resp.StatusCode, responseContent)
	}

	return nil
}

func (af PtAF) CreateRole(request Role) (string, error) {

	body := strings.NewReader(util.StructToJson(request))
	req, _ := http.NewRequest("POST", af.url+"auth/roles", body)

	resp, err := af.doRequest(*req)

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	responseContent := string(response)

	result := &CreateRoleResponse{}

	err = util.JsonToStruct(responseContent, result)
	if err != nil {
		return "", err
	}

	return result.Id, nil
}

func (af PtAF) GetRoles() []Role {

	path := "roles.json"

	if !util.FileExist(path) {
		return nil
	}

	file := util.ReadStringFromPath(path)

	var settings RolesSettings
	err := json.Unmarshal([]byte(file), &settings)
	if err != nil {
		panic(err)
	}

	return settings.Roles
}
