package vault

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/hashicorp/vault/api"
	"github.com/ibm/vault-cli/pkg/secretservice"
	"github.com/mitchellh/go-homedir"
)

type vaultservice struct {
	Client *api.Client
}

// NewVaultService should return a pointer to a vaultservice client
func NewVaultService() secretservice.SecretService {
	return &vaultservice{}
}

// SetClient should return a pointer to a vaultservice client
func (vs *vaultservice) SetClient(c *api.Client) {
	vs.Client = c
}

// SetClient should return a pointer to a vaultservice client
func (vs *vaultservice) GetClient() *api.Client {
	return vs.Client
}

// Delete is to satisfy a lint error for this interface
func (vs *vaultservice) Delete(path string) (*api.Secret, error) {
	return vs.Client.Logical().Delete(path)
}

// List is to satisfy a lint error for this interface
func (vs *vaultservice) List(path string) (*api.Secret, error) {
	return vs.Client.Logical().List(path)
}

// Read is to satisfy a lint error for this interface
func (vs *vaultservice) Read(path string) (*api.Secret, error) {
	return vs.Client.Logical().Read(path)
}

func (vs *vaultservice) ReadWithData(path string, data map[string][]string) (*api.Secret, error) {
	return vs.Client.Logical().ReadWithData(path, data)
}

// Write is to satisfy a lint error for this interface
func (vs *vaultservice) Write(path string, data map[string]interface{}) (*api.Secret, error) {
	return vs.Client.Logical().Write(path, data)
}

// IsKVv2 check version
func (vs *vaultservice) IsKVv2(path string) (string, bool, error) {
	mountPath, version, err := kvPreflightVersionRequest(vs.Client, path)
	if err != nil {
		return "", false, err
	}

	return mountPath, version == 2, nil
}

// kvPreflightVersionRequest do a preflight call
func kvPreflightVersionRequest(client *api.Client, path string) (string, int, error) {
	// We don't want to use a wrapping call here so save any custom value and
	// reservice after
	currentWrappingLookupFunc := client.CurrentWrappingLookupFunc()
	client.SetWrappingLookupFunc(nil)
	defer client.SetWrappingLookupFunc(currentWrappingLookupFunc)
	currentOutputCurlString := client.OutputCurlString()
	client.SetOutputCurlString(false)
	defer client.SetOutputCurlString(currentOutputCurlString)

	r := client.NewRequest("GET", "/v1/sys/internal/ui/mounts/"+path)
	resp, err := client.RawRequest(r)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		// If we get a 404 we are using an older version of vault, default to
		// version 1
		if resp != nil && resp.StatusCode == 404 {
			return "", 1, nil
		}

		return "", 0, err
	}

	secret, err := api.ParseSecret(resp.Body)
	if err != nil {
		return "", 0, err
	}
	if secret == nil {
		return "", 0, errors.New("nil response from pre-flight request")
	}
	var mountPath string
	if mountPathRaw, ok := secret.Data["path"]; ok {
		mountPath = mountPathRaw.(string)
	}
	options := secret.Data["options"]
	if options == nil {
		return mountPath, 1, nil
	}
	versionRaw := options.(map[string]interface{})["version"]
	if versionRaw == nil {
		return mountPath, 1, nil
	}
	version := versionRaw.(string)
	switch version {
	case "", "1":
		return mountPath, 1, nil
	case "2":
		return mountPath, 2, nil
	}

	return mountPath, 1, nil
}

// UserPassLogin will get a token from vault
func (vs *vaultservice) UserPassLogin(namespace, authurl, endpoint, username, password, cacert string, insecureSkipVerify bool) (*api.Secret, error) {
	client := &http.Client{}
	if cacert != "" {
		caCert, err := readFile(cacert)
		if err != nil {
			log.Fatal(err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:            caCertPool,
				InsecureSkipVerify: insecureSkipVerify,
			},
		}
	}
	values := map[string]string{"password": password}

	jsonValue, _ := json.Marshal(values)
	nsPath := ""
	if namespace != "" {
		nsPath = namespace + "/"
	}
	req, err := http.NewRequest("POST", authurl+"/v1/"+nsPath+"auth/"+endpoint+"/login/"+username, bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	jdata, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	secret := api.Secret{}
	err = json.Unmarshal([]byte(jdata), &secret)
	if err != nil {
		return nil, err
	}
	return &secret, nil
}

// CertLogin will get a token from vault
func (vs *vaultservice) CertLogin(namespace, url, endpoint, cert, key, cacert string, insecureSkipVerify bool) (*api.Secret, error) {
	cert = expandHomePath(cert)
	key = expandHomePath(key)
	clientCertKey, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	if cacert != "" {
		caCert, err := readFile(cacert)
		if err != nil {
			return nil, err
		}
		if len(caCert) == 0 {
			return nil, errors.New("could not read caCert")
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs:            caCertPool,
					Certificates:       []tls.Certificate{clientCertKey},
					InsecureSkipVerify: insecureSkipVerify,
				},
			},
		}
	} else {
		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					Certificates:       []tls.Certificate{clientCertKey},
					InsecureSkipVerify: insecureSkipVerify,
				},
			},
		}
	}

	req, err := http.NewRequest("POST", url+"/v1/auth/"+endpoint+"/login", nil)
	if err != nil {
		return nil, err
	}
	if namespace != "" {
		req.Header.Add("X-Vault-Namespace", namespace)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	jdata, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	secret := api.Secret{}
	err = json.Unmarshal([]byte(jdata), &secret)
	if err != nil {
		return nil, err
	}
	if secret.Auth == nil || secret.Auth.ClientToken == "" {
		return nil, fmt.Errorf("could not get token: body:%v", string(jdata))
	}
	return &secret, nil
}

// AppRoleLogin will get a token from vault
func (vs *vaultservice) AppRoleLogin(namespace, authurl, endpoint, roleID, secretID, cacert string, insecureSkipVerify bool) (*api.Secret, error) {
	client := &http.Client{}
	if cacert != "" {
		caCert, err := readFile(cacert)
		if err != nil {
			log.Fatal(err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:            caCertPool,
				InsecureSkipVerify: insecureSkipVerify,
			},
		}
	}
	values := map[string]string{"role_id": roleID, "secret_id": secretID}

	jsonValue, _ := json.Marshal(values)
	nsPath := ""
	if namespace != "" {
		nsPath = namespace + "/"
	}
	req, err := http.NewRequest("POST", authurl+"/v1/"+nsPath+"auth/"+endpoint+"/login", bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	jdata, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	secret := api.Secret{}
	err = json.Unmarshal([]byte(jdata), &secret)
	if err != nil {
		return nil, err
	}
	return &secret, nil
}

// getHomeDir returns the home dir
func getHomeDir() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}
	return home, nil
}

// expandHomePath translate home dir
func expandHomePath(path string) string {
	if path != "" && path[:1] == "~" {
		home, err := getHomeDir()
		if err != nil {
			return ""
		}
		return home + path[1:]
	}
	return path
}

// readFile expands home dir
func readFile(filename string) ([]byte, error) {
	fn := expandHomePath(filename)
	return ioutil.ReadFile(fn)
}
