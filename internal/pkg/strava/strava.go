package strava

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/adiazny/easy-strava-upload/internal/pkg/store"
	"github.com/sirupsen/logrus"
)

const (
	activitiesEndpoint = "https://www.strava.com/api/v3/activities"
)

type Provider struct {
	Log          *logrus.Entry
	ProviderName string
	Config       *Config
	HTTPClient   *http.Client
	Redis        *store.Redis
}

// Config Struct
type Config struct {
	StravaClientID     string
	StravaClientSecret string
	StravaRefreshToken string
	Scopes             []string
}

func (provider *Provider) PostActivity(uiReq *http.Request) error {

	client := &http.Client{}

	//Parse Form params coming in from UI
	uiReq.ParseForm()
	urlValues := uiReq.Form

	req, err := http.NewRequest(http.MethodPost, activitiesEndpoint, strings.NewReader(urlValues.Encode()))
	if err != nil {
		provider.Log.Infof("Error creating HTTP request %s: %v", activitiesEndpoint, err)
		return err
	}

	accessToken, refreshToken, err := provider.GetTokens()
	if err != nil {
		provider.Log.Infof("Error retrieving refresh token: %v", err)
		return err
	}

	if accessToken == "" {
		accessToken, err = provider.RefreshToken(refreshToken)
		if err != nil {
			return err
		}
	}

	// check access token expiration
	isTokenExpired, err := provider.checkAccessTokenExpired()
	if err != nil {
		provider.Log.Infof("Error checking access token: %v", err)
		return err
	}

	if isTokenExpired {
		// if expired, refresh token
		accessToken, err = provider.RefreshToken(refreshToken)
		if err != nil {
			return err
		}
	}

	bearer := fmt.Sprintf("Bearer %s", accessToken)
	req.Header.Add("Authorization", bearer)

	resp, err := client.Do(req)
	if err != nil {
		provider.Log.Infof("Error making HTTP POST request to Strava /activities: %v", err)
		return err

	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		provider.Log.Infof("HTTP POST to /activites returned reponse code and status: %d %s: ", resp.StatusCode, resp.Status)
		return err
	}

	return nil
}

func NewProvider(log *logrus.Entry, name string, c *Config, rdb *store.Redis) *Provider {
	p := new(Provider)
	p.Log = log
	p.ProviderName = name
	p.Config = c
	p.HTTPClient = http.DefaultClient
	p.Redis = rdb
	return p
}
