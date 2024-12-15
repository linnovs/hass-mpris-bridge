package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/charmbracelet/log"
	"github.com/linnovs/hass-mpris-bridge/internal/hassmessage"
)

const (
	headerAuthorization = "authorization"
	headerContentType   = "content-type"
	bearerTokenFmt      = "Bearer %s"
	contentTypeJSON     = "application/json"
)

func getInitState(bdg *bridge) (success bool) {
	apiUrl := bdg.hassURL.JoinPath("/api/states")

	req, err := http.NewRequest(http.MethodGet, apiUrl.String(), nil)
	if err != nil {
		log.Error("create HASS API request failed", "err", err)
		return false
	}

	req.Header.Set(headerAuthorization, fmt.Sprintf(bearerTokenFmt, os.Getenv(envkeyToken)))
	req.Header.Set(headerContentType, contentTypeJSON)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("failed to make request to HASS API", "err", err)
		return false
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("read response body from HASS API failed", "err", err)
		return false
	}

	var states []hassmessage.State

	if err := json.Unmarshal(body, &states); err != nil {
		log.Error("unmarshal response body from HASS API failed", "err", err)
		return false
	}

	for _, state := range states {
		if !state.IsMediaPlayer() || !state.IsMusicPlayer() {
			continue
		}

		log.Debug(
			"state from api",
			"entity", state.EntityID,
			"state", state.State,
			"attributes", string(state.Attributes),
		)

		bdg.update(state)
	}

	return true
}
