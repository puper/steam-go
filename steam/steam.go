package steam

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"

	"go.uber.org/ratelimit"
)

const defaultUserAgent = "github.com/Jleagle/steam-go"

var statusCodes = map[int]string{
	400: "steam-go: (400) please verify that all required parameters are being sent.",
	401: "steam-go: (401) access is denied. retrying will not help. please verify your key= parameter.",
	403: "steam-go: (403) access is denied. retrying will not help. please verify your key= parameter.",
	404: "steam-go: (404) the api requested does not exists.",
	405: "steam-go: (405) this api has been called with a the wrong http method like get or push.",
	429: "steam-go: (429) you are being rate limited.",
	500: "steam-go: (500) an unrecoverable error has occurred, please try again. if this continues to persist then please post to the steamworks developer discussion with additional details of your request.",
	503: "steam-go: (503) server is temporarily unavailable, or too busy to respond. please wait and try again later.",
}

type Steam struct {
	Key        string      // api key
	LogChannel chan string // channel to return call URLs
	UserAgent  string

	rateLimit     bool
	apiThrottle   ratelimit.Limiter
	storeThrottle ratelimit.Limiter
}

func (s *Steam) SetRateLimit(rate int) {

	s.rateLimit = true
	s.apiThrottle = ratelimit.New(rate)
	s.storeThrottle = ratelimit.New(rate)
}

func (s Steam) getFromAPI(path string, query url.Values) (bytes []byte, err error) {

	if s.rateLimit {
		s.apiThrottle.Take()
	}

	query.Add("format", "json")
	query.Add("key", s.Key)

	path = "https://api.steampowered.com/" + path + "?" + query.Encode()

	if s.LogChannel != nil {
		s.LogChannel <- path
	}

	client := &http.Client{}

	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return bytes, err
	}

	if s.UserAgent == "" {
		s.UserAgent = defaultUserAgent
	}

	req.Header.Set("User-Agent", s.UserAgent)

	response, err := client.Do(req)
	if err != nil {
		return bytes, err
	}
	defer response.Body.Close()

	// Handle errors
	if response.StatusCode != 200 {
		if val, ok := statusCodes[response.StatusCode]; ok {
			return bytes, Error{val, response.StatusCode}
		} else {
			return bytes, errors.New("steam: something went wrong")
		}
	}

	//
	return ioutil.ReadAll(response.Body)
}

func (s Steam) getFromStore(path string, query url.Values) (bytes []byte, err error) {

	if s.rateLimit {
		s.storeThrottle.Take()
	}

	path = "https://store.steampowered.com/" + path + "?" + query.Encode()

	if s.LogChannel != nil {
		s.LogChannel <- path
	}

	client := &http.Client{}

	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return bytes, err
	}

	if s.UserAgent == "" {
		s.UserAgent = defaultUserAgent
	}

	req.Header.Set("User-Agent", s.UserAgent)

	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	return ioutil.ReadAll(response.Body)
}

type Error struct {
	err  string
	code int
}

func (e Error) Error() string {
	return e.err
}

func (e Error) Code() int {
	return e.code
}
