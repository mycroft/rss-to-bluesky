package bluesky

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type Session struct {
	DID        string `json:"did"`
	AccessJWT  string `json:"accessJwt"`
	RefreshJWT string `json:"refreshJwt"`
	Active     bool   `json:"active"`
	Handle     string `json:"handle"`
	Email      string `json:"email"`
}

func (bs *BlueskyClient) CheckSession() error {
	var session Session

	if bs.Ready {
		return nil
	}

	// Find out if there is a valid access token in database
	encodedSession, err := bs.DB.Get("session")
	if err != nil {
		panic(err)
	}

	if len(encodedSession) > 0 {
		err = json.Unmarshal(encodedSession, &session)
		if err != nil {
			return fmt.Errorf("error unmarshaling response body: %v", err)
		}

		session, err = bs.RefreshSession(session.RefreshJWT)
		if err != nil {
			log.Printf("error refreshing session: %v", err)
			session = Session{}
		}
	}

	if session.AccessJWT == "" {
		session, err = bs.GetSession(os.Getenv("BLUESKY_USER"), os.Getenv("BLUESKY_PASS"))
		if err != nil {
			return fmt.Errorf("error getting session: %v", err)
		}
	}

	encodedSession, err = json.Marshal(session)
	if err != nil {
		return fmt.Errorf("error marshaling session: %v", err)
	}

	err = bs.DB.Set("session", encodedSession)
	if err != nil {
		return fmt.Errorf("error saving session: %v", err)
	}

	bs.Session = session
	bs.Ready = true

	return nil
}

func (bs *BlueskyClient) RefreshSession(refrestJWT string) (Session, error) {
	var session Session
	endpoint := "/xrpc/com.atproto.server.refreshSession"

	bs.Session.AccessJWT = refrestJWT

	encodedSession, err := bs.query("POST", endpoint, nil)
	if err != nil {
		return Session{}, fmt.Errorf("error querying bluesky.social: %v", err)
	}

	err = json.Unmarshal(encodedSession, &session)
	if err != nil {
		return Session{}, fmt.Errorf("error unmarshaling response body: %v", err)
	}

	return session, nil
}

func (bs *BlueskyClient) GetSession(handle, password string) (Session, error) {
	endpoint := "/xrpc/com.atproto.server.createSession"

	// Create a new session on bluesky.social
	params := map[string]string{
		"identifier": handle,
		"password":   password,
	}

	resp, err := bs.query("POST", endpoint, params)
	if err != nil {
		return Session{}, fmt.Errorf("error querying bluesky.social: %v", err)
	}

	var responsePayload Session
	err = json.Unmarshal(resp, &responsePayload)
	if err != nil {
		return Session{}, fmt.Errorf("error unmarshaling response body: %v", err)
	}

	return responsePayload, nil
}

func (bs *BlueskyClient) GetUser() error {
	endpoint := "/xrpc/app.bsky.actor.getProfile"

	params := map[string]string{
		"actor": bs.Session.DID,
	}

	resp, err := bs.query("GET", endpoint, params)
	if err != nil {
		return fmt.Errorf("error querying bluesky.social: %v", err)
	}

	fmt.Println(string(resp))

	return nil
}
