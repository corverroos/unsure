package internal

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"sort"
	"time"

	"github.com/corverroos/unsure/engine"
)

type Match struct {
	ID        int64
	Team      string
	Players   int
	Status    MatchStatus
	Summary   MatchSummary
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Round struct {
	ID        int64
	MatchID   int64
	Index     int64
	Team      string
	Status    RoundStatus
	State     RoundState
	Error     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type RoundState struct {
	Players []RoundPlayerState
}

func (r RoundState) Value() (driver.Value, error) {
	return json.MarshalIndent(r, "", " ")
}

func (r *RoundState) Scan(src interface{}) error {
	var s sql.NullString
	if err := s.Scan(src); err != nil {
		return err
	}
	*r = RoundState{}
	if !s.Valid {
		return nil
	}
	return json.Unmarshal([]byte(s.String), r)
}

func (r RoundState) GetPlayer(player string) (int, RoundPlayerState, bool) {
	for i, m := range r.Players {
		if m.Name == player {
			return i, m, true
		}
	}
	return 0, RoundPlayerState{}, false
}

func (r RoundState) GetSubmitOrder() []RoundPlayerState {
	var res []RoundPlayerState
	for _, m := range r.Players {
		if !m.Included {
			continue
		}
		res = append(res, m)
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].Rank < res[j].Rank
	})

	return res
}

func (r RoundState) GetTotal(player string) int {
	var res int
	for _, m := range r.Players {
		res += m.Parts[player]
	}
	return res
}

type RoundPlayerState struct {
	Name      string
	Rank      int
	Parts     map[string]int
	Included  bool
	Collected bool
	Submitted bool
}

type MatchStatus int

func (m MatchStatus) Enum() int {
	return int(m)
}

func (m MatchStatus) ShiftStatus() {}

func (m MatchStatus) ReflexType() int {
	return int(m)
}

const (
	MatchStatusUnknown  RoundStatus = 0
	MatchStatusStarted  MatchStatus = 1
	MatchStatusEnded    MatchStatus = 2
	matchStatusSentinel MatchStatus = 3 // This may not
)

type MatchSummary struct {
	RoundsSuccess int
	RoundsFailed  int
	Duration      time.Duration
}

func (ms *MatchSummary) Value() (driver.Value, error) {
	return json.MarshalIndent(ms, "", " ")
}

func (ms *MatchSummary) Scan(src interface{}) error {
	var s sql.NullString
	if err := s.Scan(src); err != nil {
		return err
	}
	*ms = MatchSummary{}
	if !s.Valid {
		return nil
	}
	return json.Unmarshal([]byte(s.String), ms)
}

//go:generate stringer -type=RoundStatus -trimprefix=RoundStatus

type RoundStatus int

func (m RoundStatus) Enum() int {
	return int(m)
}

func (m RoundStatus) Valid() bool {
	return m > RoundStatusUnknown && m < roundStatusSentinel
}

func (m RoundStatus) ShiftStatus() {}

// ThisOrNext returns true if m2 is equaled to m or it's next non-failed state.
func (m RoundStatus) ThisOrNext(m2 RoundStatus) bool {
	if !m2.Valid() || m2 == RoundStatusFailed {
		return false
	}
	return m == m2 || (RoundStatus(int(m)+1) == m2)
}

func (m RoundStatus) ReflexType() int {
	return engine.RoundEventOffset + int(m) // Hack to combine Match and Round events in same table.
}

const (
	RoundStatusUnknown   RoundStatus = 0
	RoundStatusJoin      RoundStatus = 1
	RoundStatusJoined    RoundStatus = 2
	RoundStatusCollect   RoundStatus = 3
	RoundStatusCollected RoundStatus = 4
	RoundStatusSubmit    RoundStatus = 5
	RoundStatusSubmitted RoundStatus = 6
	RoundStatusSuccess   RoundStatus = 7
	RoundStatusFailed    RoundStatus = 8
	roundStatusSentinel  RoundStatus = 9
)
