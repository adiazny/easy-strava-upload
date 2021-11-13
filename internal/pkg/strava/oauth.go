package strava

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	stravaTokenURL  = "https://www.strava.com/oauth/token"
	grantType       = "refresh_token"
	AthleteID       = "1232742"
	AthleteUsername = "alan_diaz"
)

type AthleteAccessInfo struct {
	ID           string   `json:"id"`
	Username     string   `json:"username"`
	RefreshToken string   `json:"refresh_token"`
	AccessToken  string   `json:"access_token"`
	ExpiresAt    int      `json:"expires_at"`
	ExpiresIn    int      `json:"expires_in"`
	Scopes       []string `json:"scopes"`
}

type APIAccess struct {
	StravaClientID     string `json:"client_id"`
	StravaClientSecret string `json:"client_secret"`
}

type stravaRefreshResponse struct {
	TokenType    string `json:"token_type"`
	AccesToken   string `json:"access_token"`
	ExpiresAt    int    `json:"expires_at"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// RefreshToken refreshes the OAUTH access token
func (provider *Provider) RefreshToken(rt string) (accessToken string, err error) {
	var formData url.Values

	formData = url.Values{
		"grant_type":    {grantType},
		"client_id":     {provider.Config.StravaClientID},
		"client_secret": {provider.Config.StravaClientSecret},
		"refresh_token": {rt},
	}

	encodedFormData := formData.Encode()

	req, err := http.NewRequest(http.MethodPost, stravaTokenURL, strings.NewReader(encodedFormData))
	if err != nil {
		provider.Log.Infof("Error creating HTTP Request %s: %v", stravaTokenURL, err)
		return accessToken, err
	}

	resp, err := provider.HTTPClient.Do(req)
	if err != nil {
		provider.Log.Infof("Error making HTTP POST to Strava OAuth /token: %v", err)
		return accessToken, err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		provider.Log.Fatal(err)
	}

	var refreshResp stravaRefreshResponse

	json.Unmarshal(b, &refreshResp)

	athleteAccess := &AthleteAccessInfo{
		ID:           AthleteID,
		Username:     AthleteUsername,
		RefreshToken: refreshResp.RefreshToken,
		AccessToken:  refreshResp.AccesToken,
		ExpiresAt:    refreshResp.ExpiresAt,
		ExpiresIn:    refreshResp.ExpiresIn,
	}

	athleteAccessInfoJSON, err := json.Marshal(athleteAccess)
	if err != nil {
		return accessToken, fmt.Errorf("Error marshling value %v to JSON", athleteAccess)
	}

	err = provider.Redis.Store(AthleteID, athleteAccessInfoJSON)
	if err != nil {
		return accessToken, err
	}

	accessToken = refreshResp.AccesToken

	return
}

func (provider *Provider) GetTokens() (accessToken, refreshToken string, err error) {
	athleteAccessInfo, err := provider.getAthleteAccessInfo()
	if err != nil {
		return "", "", err
	}

	return athleteAccessInfo.AccessToken, athleteAccessInfo.RefreshToken, nil
}

func (provider *Provider) getAthleteAccessInfo() (athleteAccess *AthleteAccessInfo, err error) {
	val, err := provider.Redis.Client.Get(context.Background(), AthleteID).Result()

	if err != nil {
		return athleteAccess, err
	}

	b := []byte(val)
	json.Unmarshal(b, &athleteAccess)

	return athleteAccess, nil
}

func (provider *Provider) checkAccessTokenExpired() (expired bool, err error) {
	athleteAccessInfo, err := provider.getAthleteAccessInfo()
	if err != nil {
		return expired, err
	}

	timeNow := time.Now()
	timeSecs := timeNow.Unix()
	if int64(athleteAccessInfo.ExpiresAt) < timeSecs {
		return true, nil
	}

	return
}
