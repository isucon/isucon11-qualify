package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

type Portal struct {
	Endpoint string
	Token    string
}

type EnvCheckInfo struct {
	AMI []string `json:"ami_ids"`
	AZ  string   `json:"az_id"`
}

func LoadPortalCredentials() (*Portal, error) {
	var credentials struct {
		Dev   bool   `json:"dev"`
		Token string `json:"token"`
	}
	f, err := os.Open("/opt/isucon-env-checker/portal_credentials.json")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&credentials); err != nil {
		return nil, err
	}
	endpoint := "https://portal.isucon.net"
	if credentials.Dev {
		endpoint = "https://portal-dev.isucon.net"
	}
	return &Portal{
		Endpoint: endpoint,
		Token:    credentials.Token,
	}, nil
}

func (p *Portal) GetInfo(name string) (EnvCheckInfo, error) {
	q := make(url.Values)
	q.Set("name", name)
	q.Set("token", p.Token)
	req, err := http.NewRequest("GET", p.Endpoint+"/api/env_check_info?"+q.Encode(), nil)
	if err != nil {
		return EnvCheckInfo{}, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return EnvCheckInfo{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(res.Body)
		return EnvCheckInfo{}, fmt.Errorf("http status error: %d (%s)", res.StatusCode, string(msg))
	}

	var info EnvCheckInfo
	if err := json.NewDecoder(res.Body).Decode(&info); err != nil {
		return EnvCheckInfo{}, err
	}

	return info, nil
}

func (p *Portal) SendResult(r Result) error {
	body, err := json.Marshal(struct {
		Token        string `json:"token"`
		Name         string `json:"name"`
		Passed       bool   `json:"passed"`
		IPAddress    string `json:"ip_address"`
		Message      string `json:"message"`
		AdminMessage string `json:"admin_message"`
		RawData      string `json:"raw_data"`
	}{
		Token:        p.Token,
		Name:         r.Name,
		Passed:       r.Passed,
		IPAddress:    r.IPAddress,
		Message:      r.Message,
		AdminMessage: r.AdminMessage,
		RawData:      r.RawData,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", p.Endpoint+"/api/env_checks", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		msg, _ := io.ReadAll(res.Body)
		return fmt.Errorf("http status error: %d (%s)", res.StatusCode, string(msg))
	}

	io.Copy(io.Discard, res.Body)
	return nil
}
