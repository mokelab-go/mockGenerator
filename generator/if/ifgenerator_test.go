package ifgenerator

import "testing"
import "fmt"

func Test_generator_Generate(t *testing.T) {
	src := `package moke
import (
	m "../"
	"fmt"
)

// +mock	
// MokeRepo is a repository
type MokeRepo interface {
	NoResult()
	NoArgs() error
	Moke(a, b string, c int) (string, error)
	CustomError() m.MyError
}
`
	expected := `package moke

import (
	m "../"
	"fmt"
)

type MockMokeRepo struct {
	NoArgsResult0      error
	MokeResult0        string
	MokeResult1        error
	CustomErrorResult0 m.MyError
}

func (m *MockMokeRepo) NoResult()                               { return }
func (m *MockMokeRepo) NoArgs() error                           { return m.NoArgsResult0 }
func (m *MockMokeRepo) Moke(a, b string, c int) (string, error) { return m.MokeResult0, m.MokeResult1 }
func (m *MockMokeRepo) CustomError() m.MyError                  { return m.CustomErrorResult0 }
`

	g := New()
	out, err := g.Generate(src)
	if err != nil {
		t.Errorf("Failed to generate %s", err)
		return
	}
	fmt.Printf("%s", out)
	if out != expected {
		t.Errorf("Wrong output %s len=%d - %d", out, len(out), len(expected))
		return
	}
}
