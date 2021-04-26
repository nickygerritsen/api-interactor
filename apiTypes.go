package interactor

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type (
	// ApiType is an interface used for objects that interact with the API.
	ApiType interface {
		FromJSON([]byte) (ApiType, error)
		fmt.Stringer

		Path() string
		Generate() ApiType
		InContest() bool
	}

	// Submittable is an ApiType that can be submitted to the API. TODO decide on whether to merge the interfaces
	Submittable interface {
		ApiType
	}

	// ApiTime is a time.Time which marshals to and from the format used in the CCS Api
	ApiTime time.Time

	// ApiRelTime is a time.Duration which marshals to and from the format used in the CCS Api
	ApiRelTime time.Duration

	// TODO add omitempty to appropriate keys

	Contest struct {
		Id         string     `json:"id"`
		Name       string     `json:"name"`
		FormalName string     `json:"formal_name"`
		StartTime  ApiTime    `json:"start_time"`
		Duration   ApiRelTime `json:"duration"`
	}

	Problem struct {
		Id      string `json:"id"`
		Label   string `json:"label"`
		Name    string `json:"name"`
		Ordinal int    `json:"ordinal"`
	}

	Submission struct {
		Id          string     `json:"id"`
		LanguageId  string     `json:"language_id"`
		ContestTime ApiRelTime `json:"contest_time"`
		TeamId      string     `json:"team_id"`
		ProblemId   string     `json:"problem_id"`
		ExternalId  string     `json:"external_id"`
		EntryPoint  string     `json:"entry_point"`
		File        string     `json:"file"`
	}

	Clarification struct {
		Id          string     `json:"id,omitempty"`
		FromTeamId  string     `json:"from_team_id,omitempty"`
		ToTeamId    string     `json:"to_team_id,omitempty"`
		ReplyToId   string     `json:"reply_to_id,omitempty"`
		ProblemId   string     `json:"problem_id"`
		Text        string     `json:"text"`
		Time        *ApiTime   `json:"time,omitempty"`
		ContestTime ApiRelTime `json:"contest_time,omitempty"`
	}

	Identifier string
)

// -- Contest implementation

func (c Contest) FromJSON(data []byte) (ApiType, error) {
	err := json.Unmarshal(data, &c)
	return c, err
}

func (c Contest) String() string {
	// TODO format the starttime and duration
	return fmt.Sprintf(`
         id: %v
       name: %v
formal name: %v
 start time: %v
   duration: %v
`, c.Id, c.Name, c.FormalName, c.StartTime, c.Duration)
}

func (c Contest) InContest() bool {
	return false
}

func (c Contest) Path() string {
	return "contests"
}

func (c Contest) Generate() ApiType {
	return Contest{}
}

// -- Problem implementation

func (p Problem) FromJSON(data []byte) (ApiType, error) {
	err := json.Unmarshal(data, &p)
	return p, err
}

func (p Problem) String() string {
	return fmt.Sprintf(`
         id: %v
      label: %v
       name: %v
    ordinal: %v
`, p.Id, p.Label, p.Name, p.Ordinal)
}

func (p Problem) Path() string {
	return "problems"
}

func (p Problem) InContest() bool {
	return true
}

func (p Problem) Generate() ApiType {
	return Problem{}
}

// -- Submission implementation

func (s Submission) FromJSON(data []byte) (ApiType, error) {
	err := json.Unmarshal(data, &s)
	return s, err
}

func (s Submission) InContest() bool {
	return true
}

func (s Submission) Path() string {
	return "submissions"
}

func (s Submission) Generate() ApiType {
	return Submission{}
}

func (s Submission) String() string {
	return fmt.Sprintf(`
          id: %v
 language id: %v
contest time: %v
     team id: %v
  problem id: %v
 external id: %v
 entry point: %v
`, s.Id, s.LanguageId, s.ContestTime, s.TeamId, s.ProblemId, s.ExternalId, s.EntryPoint)
}

// -- Clarification implementation

func (c Clarification) FromJSON(data []byte) (ApiType, error) {
	err := json.Unmarshal(data, &c)
	return c, err
}

func (c Clarification) InContest() bool {
	return true
}

func (c Clarification) Path() string {
	return "clarifications"
}

func (c Clarification) Generate() ApiType {
	return Clarification{}
}

func (c Clarification) String() string {
	return fmt.Sprintf(`
 from team id: %v
   to team id: %v
  reply to id: %v
   problem id: %v
         text: %v
         time: %v
 contest time: %v
`, c.FromTeamId, c.ToTeamId, c.ReplyToId, c.ProblemId, c.Text, c.Time, c.ContestTime)
}

// -- ApiTime implementation

func (a *ApiTime) UnmarshalJSON(b []byte) (err error) {
	data := strings.Trim(string(b), "\"")

	if data == "null" {
		*a = ApiTime(time.Time{})
		return
	}

	// All possible time formats we support
	var supportedTimeFormats = []string{
		// time.RFC3999 also accepts milliseconds, even though it is not officially stated
		time.RFC3339,
		// time.RFC3999 but then without the minutes of the timezone
		"2006-01-02T15:04:05Z07",
	}
	for _, supportedTimeFormat := range supportedTimeFormats {
		if t, err := time.Parse(supportedTimeFormat, data); err == nil {
			*a = ApiTime(t)
			return nil
		}
	}

	return fmt.Errorf("can not format date: %s", data)
}

// -- ApiRelTime implementation

func (a *ApiRelTime) UnmarshalJSON(b []byte) (err error) {
	data := strings.Trim(string(b), "\"")
	if data == "null" {
		*a = 0
		return
	}
	re := regexp.MustCompile("(-?[0-9]{1,2}):([0-9]{2}):([0-9]{2})(.([0-9]{3}))?")
	sm := re.FindStringSubmatch(data)
	h, err := strconv.ParseInt(sm[1], 10, 64)
	if err != nil {
		return err
	}

	m, err := strconv.ParseInt(sm[2], 10, 64)
	if err != nil {
		return err
	}

	s, err := strconv.ParseInt(sm[3], 10, 64)
	if err != nil {
		return err
	}

	var ms int64 = 0
	if sm[5] != "" {
		ms, err = strconv.ParseInt(sm[5], 10, 64)
		if err != nil {
			return err
		}
	}

	*a = ApiRelTime(time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(s)*time.Second + time.Duration(ms)*time.Millisecond)

	return
}

// -- ApiRelTime implementation

func (a ApiRelTime) String() string {
	return time.Duration(a).String()
}

func (a ApiRelTime) Duration() time.Duration {
	return time.Duration(a)
}

func (a ApiTime) Time() time.Time {
	return time.Time(a)
}

func (a ApiTime) String() string {
	return time.Time(a).String()
}

// -- Identifier implementation

func (i *Identifier) UnmarshalJSON(bts []byte) error {
	// It is expected to be a string, possible embedded in quotes
	*i = Identifier(strings.Trim(string(bts), "\"'"))
	return nil
}
