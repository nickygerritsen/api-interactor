package interactor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func (i inter) Contests() ([]Contest, error) {
	obj, err := i.GetObjects(Contest{})
	if err != nil {
		return nil, fmt.Errorf("could not retrieve contests; %w", err)
	}

	// obj should be a slice of Contest, cast to it to slice of Contest
	ret := make([]Contest, len(obj))
	for k, v := range obj {
		vv, ok := v.(Contest)
		if !ok {
			return ret, fmt.Errorf("unexpected type found, expected contest, got: %T", v)
		}

		ret[k] = vv
	}

	return ret, nil
}

func (i inter) ContestById(contestId string) (c Contest, err error) {
	// Retrieve all contests and check whether the contest exists, TODO decide on whether to optimize
	contests, err := i.Contests()
	if err != nil {
		return c, fmt.Errorf("could not retrieve contest")
	}

	for _, v := range contests {
		if v.Id == contestId {
			return v, nil
		}
	}

	return c, errNotFound
}

func (i inter) Problems() ([]Problem, error) {
	obj, err := i.GetObjects(Problem{})
	if err != nil {
		return nil, fmt.Errorf("could not retrieve problems; %w", err)
	}

	// obj should be a slice of Problem, cast to it to slice of Problem
	ret := make([]Problem, len(obj))
	for k, v := range obj {
		vv, ok := v.(Problem)
		if !ok {
			return ret, fmt.Errorf("unexpected type found, expected problem, got: %T", v)
		}

		ret[k] = vv
	}

	return ret, nil
}

func (i inter) ProblemById(problemId string) (p Problem, err error) {
	obj, err := i.GetObject(p, problemId)
	if err != nil {
		return p, fmt.Errorf("could not retrieve problem; %w", err)
	}

	vv, ok := obj.(Problem)
	if !ok {
		return p, fmt.Errorf("unexpected type found, expected problem, got: %T", obj)
	}

	p = vv
	return
}

func (i inter) Submissions() ([]Submission, error) {
	obj, err := i.GetObjects(Submission{})
	if err != nil {
		return nil, fmt.Errorf("could not retrieve submissions; %w", err)
	}

	// obj should be a slice of Submission, cast to it to slice of Submission
	ret := make([]Submission, len(obj))
	for k, v := range obj {
		vv, ok := v.(Submission)
		if !ok {
			return ret, fmt.Errorf("unexpected type found, expected submission, got: %T", v)
		}

		ret[k] = vv
	}

	return ret, nil
}

func (i inter) SubmissionById(submissionId string) (s Submission, err error) {
	obj, err := i.GetObject(s, submissionId)
	if err != nil {
		return s, fmt.Errorf("could not retrieve submission; %w", err)
	}

	vv, ok := obj.(Submission)
	if !ok {
		return s, fmt.Errorf("unexpected type found, expected submission, got: %T", obj)
	}

	s = vv
	return
}

func (i inter) Languages() ([]Language, error) {
	obj, err := i.GetObjects(Language{})
	if err != nil {
		return nil, fmt.Errorf("could not retrieve languages; %w", err)
	}

	// obj should be a slice of Language, cast to it to slice of Language
	ret := make([]Language, len(obj))
	for k, v := range obj {
		vv, ok := v.(Language)
		if !ok {
			return ret, fmt.Errorf("unexpected type found, expected language, got: %T", v)
		}

		ret[k] = vv
	}

	return ret, nil
}

func (i inter) LanguageById(languageId string) (l Language, err error) {
	obj, err := i.GetObject(l, languageId)
	if err != nil {
		return l, fmt.Errorf("could not retrieve language; %w", err)
	}

	vv, ok := obj.(Language)
	if !ok {
		return l, fmt.Errorf("unexpected type found, expected language, got: %T", obj)
	}

	l = vv
	return
}

func (i inter) PostClarification(problemId, text string) (Identifier, error) {
	return i.postToId(i.contestPath("clarifications"), Clarification{
		ProblemId: problemId,
		Text:      text,
	})
}

func (i inter) PostSubmission(problemId, languageId, entrypoint string, files LocalFileReference) (Identifier, error) {
	return i.postToId(i.contestPath("submissions"), Submission{
		ProblemId:  problemId,
		LanguageId: languageId,
		EntryPoint: entrypoint,
		Files: []FileReference{
			{
				Mime: "application/zip",
				Data: files,
			},
		},
	})
}

func (i inter) Submit(s Submittable) (Identifier, error) {
	return i.postToId(i.contestPath(s.Path()), s)
}

func (i inter) GetObject(interactor ApiType, id string) (ApiType, error) {
	objs, err := i.retrieve(interactor, i.toPath(interactor)+id, true)

	if err != nil {
		return nil, fmt.Errorf("could not retrieve; %w", err)
	}

	if len(objs) != 1 {
		return nil, fmt.Errorf("incorrect number of objects found, expected 1, got: %v", len(objs))
	}

	return objs[0], nil
}

func (i inter) toPath(interactor ApiType) string {
	var base string
	if interactor.InContest() {
		base = "contests/" + i.contestId + "/"
	}

	return base + interactor.Path() + "/"
}

func (i inter) GetObjects(interactor ApiType) ([]ApiType, error) {
	return i.retrieve(interactor, i.toPath(interactor), false)
}

func (i inter) retrieve(interactor ApiType, path string, single bool) ([]ApiType, error) {
	resp, err := i.Get(i.baseUrl + path)
	if err != nil {
		return nil, err
	}

	// Body is not-nil, ensure it will always be closed
	defer resp.Body.Close()

	if err := statusToError(resp.StatusCode); err != nil {
		return nil, err
	}

	// If id is not empty, only a single instance is expected to be returned
	if single {
		bts, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("could not read entire body; %w", err)
		}

		in, err := interactor.FromJSON(bts)
		return []ApiType{in}, err
	}

	// Some json should be returned, construct a decoder
	decoder := json.NewDecoder(resp.Body)

	// We read everything into a slice of
	var temp []json.RawMessage
	if err := decoder.Decode(&temp); err != nil {
		return nil, err
	}

	// Create the actual slice to return
	ret := make([]ApiType, len(temp))
	for k, v := range temp {
		// Generate a new interactor
		vv, err := interactor.FromJSON(v)
		if err != nil {
			return ret, err
		}

		ret[k] = vv
	}

	return ret, nil
}

func (i inter) postToId(path string, encodableBody Submittable) (Identifier, error) {
	var returnedId Identifier

	var buf = new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(encodableBody)
	if err != nil {
		return returnedId, fmt.Errorf("could not marshal body; %w", err)
	}

	// Post the body
	resp, err := i.Post(i.baseUrl+path, "application/json", buf)
	if err != nil {
		return returnedId, fmt.Errorf("could not post request; %w", err)
	}

	defer resp.Body.Close()

	if err := statusToError(resp.StatusCode); err != nil {
		return returnedId, err
	}

	return returnedId, json.NewDecoder(resp.Body).Decode(&returnedId)
}

func statusToError(status int) error {
	switch status {
	case http.StatusOK:
		return nil
	case http.StatusUnauthorized:
		return errUnauthorized
	case http.StatusNotFound:
		return errNotFound
	default:
		return fmt.Errorf("invalid statuscode received: %d", status)
	}
}
