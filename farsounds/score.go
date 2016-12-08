package farsounds

import (
	"container/list"

	"github.com/mitchellh/mapstructure"
)

// scoreEventLoader event loader for score script
type scoreEventDesc struct {
	// Time in seconds
	On float64

	// Action
	Action string

	// Payload
	Payload interface{}
}

/*
	Score actions
*/

// ScoreAction interface
type ScoreAction interface {
	Action(*ScorePlayer, Module, *float64)
}

// ScoreResetAction triggers reset on player
type ScoreResetAction struct{}

// Action reset
func (action *ScoreResetAction) Action(scorePlayer *ScorePlayer, module Module, time *float64) {
	scorePlayer.Reset()
	*time = 0.0
}

/*
	Send action
*/

// ScoreDelivery score message for addres
type ScoreDelivery struct {
	Message Message
	Address string
}

// ScoreSendAction send deliveries
type ScoreSendAction struct {
	// Messages to send
	Deliveries []*ScoreDelivery
}

// NewScoreSendAction new send action
func NewScoreSendAction(payload interface{}) *ScoreSendAction {
	action := ScoreSendAction{}

	if deliveryMap, ok := payload.(map[string]interface{}); ok {
		deliveries := make([]*ScoreDelivery, len(deliveryMap))

		// Index to deliveries
		deliveryIndex := 0

		// Add deliveries
		for address, message := range deliveryMap {
			delivery := ScoreDelivery{}
			delivery.Address = address
			delivery.Message = message
			deliveries[deliveryIndex] = &delivery
			deliveryIndex++
		}

		action.Deliveries = deliveries
	}

	return &action
}

// Action send
func (action *ScoreSendAction) Action(player *ScorePlayer, module Module, time *float64) {
	for _, delivery := range action.Deliveries {
		module.SendMessage(NewAddress(delivery.Address), delivery.Message)
	}
}

/*
	Score player
*/

// ScoreEvent event
type ScoreEvent struct {
	// Time in seconds when to trigger this entry
	On float64

	// Action
	Action ScoreAction
}

// Score event list
type Score struct {
	Events *list.List
}

// ScorePlayer for score
type ScorePlayer struct {
	Timestamp int64
	Score     *Score
	LastEvent *list.Element
}

// SetScore for player
func (player *ScorePlayer) SetScore(score *Score) {
	player.Score = score
	player.LastEvent = score.Events.Front()
}

// Play events, time is in seconds
func (player *ScorePlayer) Play(module Module) {
	time := float64(player.Timestamp) / module.GetSampleRate()

	player.Timestamp += int64(module.GetBufferLength())

	for player.LastEvent != nil && player.LastEvent.Value.(*ScoreEvent).On <= time {
		event := player.LastEvent.Value.(*ScoreEvent)

		player.LastEvent = player.LastEvent.Next()

		event.Action.Action(player, module, &time)
	}
}

// Reset score player
func (player *ScorePlayer) Reset() {
	player.Timestamp = 0
	player.LastEvent = player.Score.Events.Front()
}

// NewScorePlayer create new score player
func NewScorePlayer(score *Score) *ScorePlayer {
	return &ScorePlayer{
		Score:     score,
		LastEvent: score.Events.Front(),
	}
}

// LoadScore load score
func LoadScore(filePath string) (*Score, error) {
	_score, err := EvalScript(filePath, func(obj interface{}) (interface{}, error) {
		var rawEvents []*scoreEventDesc

		err := mapstructure.Decode(obj, &rawEvents)
		if err != nil {
			return nil, err
		}

		score := Score{}
		events := list.New()

		// Loop through raw events
		for _, rawEvent := range rawEvents {
			event := ScoreEvent{On: rawEvent.On}

			switch rawEvent.Action {
			case "send":
				event.Action = NewScoreSendAction(rawEvent.Payload)
			case "reset":
				event.Action = new(ScoreResetAction)
			}

			// Pushback on list
			events.PushBack(&event)
		}

		score.Events = events

		return &score, nil
	})

	if err != nil {
		return nil, err
	}

	return _score.(*Score), nil
}
