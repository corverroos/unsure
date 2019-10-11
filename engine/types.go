package engine

const RoundEventOffset = 100

//go:generate stringer -type=EventType -trimprefix=EventType

type EventType int

func (t EventType) Valid() bool {
	return t > EventTypeUnknown && t < eventTypeMatchSentinel ||
		t > eventTypeRoundMarker && t < eventTypeRoundSentinel
}

func (t EventType) ReflexType() int {
	return int(t)
}

const (
	EventTypeUnknown       EventType = 0
	EventTypeMatchStarted  EventType = 1
	EventTypeMatchEnded    EventType = 2
	eventTypeMatchSentinel EventType = 3

	eventTypeRoundMarker    EventType = RoundEventOffset
	EventTypeRoundJoin      EventType = RoundEventOffset + 1
	EventTypeRoundJoined    EventType = RoundEventOffset + 2
	EventTypeRoundCollect   EventType = RoundEventOffset + 3
	EventTypeRoundCollected EventType = RoundEventOffset + 4
	EventTypeRoundSubmit    EventType = RoundEventOffset + 5
	EventTypeRoundSubmitted EventType = RoundEventOffset + 6
	EventTypeRoundSuccess   EventType = RoundEventOffset + 7
	EventTypeRoundFailed    EventType = RoundEventOffset + 8
	eventTypeRoundSentinel  EventType = RoundEventOffset + 9
)

type CollectRoundRes struct {
	Rank    int
	Players []CollectPlayer
}

type CollectPlayer struct {
	Name string
	Part int
}
