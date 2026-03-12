package zonebus

import "fmt"

// stageSet holds all valid stage values for a warehouse zone.
type stageSet struct {
	Inbound     Stage
	Received    Stage
	Processing  Stage
	Assembly    Stage
	Calibration Stage
	QA          Stage
	Outbound    Stage
}

// Stages is the set of allowed zone stage values.
var Stages = stageSet{
	Inbound:     newStage("inbound"),
	Received:    newStage("received"),
	Processing:  newStage("processing"),
	Assembly:    newStage("assembly"),
	Calibration: newStage("calibration"),
	QA:          newStage("qa"),
	Outbound:    newStage("outbound"),
}

var stageMap = make(map[string]Stage)

// Stage represents the manufacturing lifecycle stage associated with a zone.
// Nullable — businesses without stage-based tracking leave this unset.
type Stage struct {
	name string
}

func newStage(s string) Stage {
	st := Stage{s}
	stageMap[s] = st
	return st
}

// String returns the string representation of the Stage.
func (s Stage) String() string {
	return s.name
}

// Equal returns true if the two Stages are equal.
func (s Stage) Equal(s2 Stage) bool {
	return s.name == s2.name
}

// MarshalText implements encoding.TextMarshaler.
func (s Stage) MarshalText() ([]byte, error) {
	return []byte(s.name), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (s *Stage) UnmarshalText(data []byte) error {
	parsed, err := ParseStage(string(data))
	if err != nil {
		return err
	}
	*s = parsed
	return nil
}

// ParseStage parses a string into a Stage.
func ParseStage(value string) (Stage, error) {
	st, exists := stageMap[value]
	if !exists {
		return Stage{}, fmt.Errorf("invalid stage %q", value)
	}
	return st, nil
}
