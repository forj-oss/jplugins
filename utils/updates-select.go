package utils

import (
	"fmt"
)

// UpdatesSelect is a choice matrix framework
type UpdatesSelect struct {
	checkers map[string]func() bool
	choices  []UpdatesSelectChoice
	state    map[string]bool
}

type UpdatesSelectChoice struct {
	Checkers []string
	callFunc func(choice UpdatesSelectChoice, state map[string]bool) error
	Choice   string
	Index    int
}

// NewUpdatesSelect creates the matrix object.
func NewUpdatesSelect() *UpdatesSelect {
	ret := new(UpdatesSelect)
	ret.checkers = make(map[string]func() bool)
	ret.choices = make([]UpdatesSelectChoice, 0, 5)
	ret.state = make(map[string]bool)
	return ret
}

// SetCheck define a checker name and test function
func (u *UpdatesSelect) SetCheck(name string, checker func() bool) {
	if u == nil {
		return
	}

	u.checkers[name] = checker
}

// SetChoice define a function to call if the check combination is found.
// First declared first executed.
func (u *UpdatesSelect) SetChoice(choice string, call func(choice UpdatesSelectChoice, states map[string]bool) error, checks ...string) {
	if u == nil {
		return
	}

	u.choices = append(u.choices, UpdatesSelectChoice{
		Checkers: checks,
		callFunc: call,
		Choice:   choice,
	})
	index := len(u.choices) - 1
	u.choices[index].Index = index
}

// Run executes the evaluations of all checkers, then loop on choices.
// The first choice true, executes the command
func (u *UpdatesSelect) Run() error {
	// Evaluate
	for name, eval := range u.checkers {
		u.state[name] = eval()
	}

	// Choose
	for _, choice := range u.choices {
		choosen := true
		for _, checker := range choice.Checkers {
			choosen = u.state[checker]
			if !choosen {
				break
			}
		}
		if choosen {
			return choice.callFunc(choice, u.state)
		}
}
	return fmt.Errorf("Unable to choose")
}
