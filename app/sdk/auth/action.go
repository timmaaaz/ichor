package auth

import "fmt"

type actionSet struct {
	Create Action
	Read   Action
	Update Action
	Delete Action
}

// Actions represents the set of actions that can be used.
var Actions = actionSet{
	Create: newAction("CREATE"),
	Read:   newAction("READ"),
	Update: newAction("UPDATE"),
	Delete: newAction("DELETE"),
}

// =============================================================================

// Set of known actions.
var actions = make(map[string]Action)

// Action represents an action in the system.
type Action struct {
	name string
}

func newAction(action string) Action {
	a := Action{action}
	actions[action] = a
	return a
}

// String returns the name of the action.
func (a Action) String() string {
	return a.name
}

// Equal provides support for the go-cmp package and testing.
func (a Action) Equal(a2 Action) bool {
	return a.name == a2.name
}

// =============================================================================

// ParseAction parses the string value and returns an action if one exists.
func ParseAction(value string) (Action, error) {
	action, exists := actions[value]
	if !exists {
		return Action{}, fmt.Errorf("invalid action %q", value)
	}

	return action, nil
}

// MustParseAction parses the string value and returns an action if one exists. If
// an error occurs the function panics.
func MustParseAction(value string) Action {
	action, err := ParseAction(value)
	if err != nil {
		panic(err)
	}

	return action
}

// ParseActionsToString takes a collection of user actions and converts them to
// a slice of string.
func ParseActionsToString(usrActions []Action) []string {
	actns := make([]string, len(usrActions))
	for i, action := range usrActions {
		actns[i] = action.String()
	}

	return actns
}

// ParseActions takes a collection of strings and converts them to a slice
// of actions.
func ParseActions(actions []string) ([]Action, error) {
	usrActions := make([]Action, len(actions))
	for i, actionStr := range actions {
		action, err := ParseAction(actionStr)
		if err != nil {
			return nil, err
		}
		usrActions[i] = action
	}

	return usrActions, nil
}
