package gh

import (
	"context"
	"strings"
)

// MockRunner is a test double for Runner.
// Called[i] and CalledStdin[i] always correspond to the same call.
// CalledStdin[i] is nil for Run calls and contains the body for RunWithStdin calls.
type MockRunner struct {
	Responses   map[string][]byte
	Errors      map[string]error
	Called      [][]string
	CalledStdin [][]byte
}

func (m *MockRunner) Run(_ context.Context, args ...string) ([]byte, error) {
	key := strings.Join(args, " ")
	m.Called = append(m.Called, args)
	m.CalledStdin = append(m.CalledStdin, nil)
	if err, ok := m.Errors[key]; ok {
		return nil, err
	}
	if resp, ok := m.Responses[key]; ok {
		return resp, nil
	}
	return nil, nil
}

func (m *MockRunner) RunWithStdin(_ context.Context, body []byte, args ...string) ([]byte, error) {
	key := strings.Join(args, " ")
	m.Called = append(m.Called, args)
	m.CalledStdin = append(m.CalledStdin, body)
	if err, ok := m.Errors[key]; ok {
		return nil, err
	}
	if resp, ok := m.Responses[key]; ok {
		return resp, nil
	}
	return nil, nil
}
