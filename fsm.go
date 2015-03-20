package fsm

import "errors"

type State string

type Guard func(subject Stater, goal State) bool

var (
  InvalidTransition = errors.New("invalid transition")
)

type Transition struct {
	Origin, Exit State
}

type Ruleset map[Transition][]Guard

func (r Ruleset) AddRule(t Transition, guards ...Guard) {
	for _, guard := range guards {
		r[t] = append(r[t], guard)
	}
}

func (r Ruleset) AddTransition(t Transition) {
	r.AddRule(t, func(subject Stater, goal State) bool {
		return subject.CurrentState() == t.Origin
	})
}

// Permitted determines if a transition is allowed
func (r Ruleset) Permitted(subject Stater, goal State) bool {
	attempt := Transition{subject.CurrentState(), goal}

	if guards, ok := r[attempt]; ok {
		outcome := make(chan bool)

		for _, guard := range guards {
			go func() {
				outcome <- guard(subject, goal)
			}()
		}

    for range guards {
      select {
        case o := <-outcome:
          if !o {
            return false
          }
      }
    }

		return true // All guards passed
	}
	return false // No rule found for the transition
}

// Stater can be passed into the FSM. The Stater reponsible for setting
// its own default state. Behavior of a Stater without a State is undefined.
type Stater interface {
	CurrentState() State
	SetState(State)
}

// Machine is a pairing of Rules and a Subject.
// The subject or rules may be changed at any time within
// the machine's lifecycle.
type Machine struct {
	Rules   *Ruleset
	Subject Stater
}

// Transition attempts to move the Subject to the Goal state.
func (m Machine) Transition(goal State) error {
	if m.Rules.Permitted(m.Subject, goal) {
		m.Subject.SetState(goal)
		return nil
	}

	return InvalidTransition
}

func New(opts ...func(*Machine)) Machine {
 var m Machine

 for _, opt := range opts { opt(&m) }

 return m
}

func WithSubject(s Stater) func(*Machine) {
  return func(m *Machine) {
    m.Subject = s
  }
}

func WithRules(r Ruleset) func(*Machine) {
  return func(m *Machine) {
    m.Rules = &r
  }
}