package farsounds

import (
	"container/list"

	"github.com/mitchellh/mapstructure"
)

// scoreEventLoader event loader for score script
type scoreEventLoader struct {
	On float64
	// Messages to send
	Send map[string]interface{}
}

// ScoreDelivery score message for addres
type ScoreDelivery struct {
	Message Message
	Address *Address
}

// ScoreEvent event
type ScoreEvent struct {
	// Time in seconds when to trigger this entry
	On float64
	// Messages to send
	Send []*ScoreDelivery
}

// Score event list
type Score struct {
	Events *list.List
}

// ScorePlayer for score
type ScorePlayer struct {
	Score     *Score
	LastEvent *list.Element
}

// SetScore for player
func (player *ScorePlayer) SetScore(score *Score) {
	player.Score = score
	player.LastEvent = score.Events.Front()
}

// Play events, time is in seconds
func (player *ScorePlayer) Play(time float64, module Module) {
	for player.LastEvent != nil && player.LastEvent.Value.(*ScoreEvent).On <= time {
		event := player.LastEvent.Value.(*ScoreEvent)

		if event.Send != nil {
			for _, delivery := range event.Send {
				module.SendMessage(delivery.Address, delivery.Message)
			}
		}

		player.LastEvent = player.LastEvent.Next()
	}
}

// Reset score player
func (player *ScorePlayer) Reset() {
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
		var rawEvents []*scoreEventLoader

		err := mapstructure.Decode(obj, &rawEvents)
		if err != nil {
			return nil, err
		}

		score := Score{}
		events := list.New()

		// Loop through raw events
		for _, rawEvent := range rawEvents {
			event := ScoreEvent{}

			// Pushback on list
			events.PushBack(&event)

			// Set On time
			event.On = rawEvent.On

			// Add messages to send array
			if rawEvent.Send != nil {
				event.Send = make([]*ScoreDelivery, len(rawEvent.Send))

				// Index to deliveries
				deliveryIndex := 0

				// Add deliveries
				for address, message := range rawEvent.Send {
					delivery := ScoreDelivery{}
					delivery.Address = NewAddress(address)
					delivery.Message = message
					event.Send[deliveryIndex] = &delivery
					deliveryIndex++
				}
			}
		}

		score.Events = events

		return &score, nil
	})

	if err != nil {
		return nil, err
	}

	return _score.(*Score), nil
}
