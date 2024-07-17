package connection

import "fmt"

type Text struct {
	Text   string `json:"text"`
	Size   int    `json:"size"`
	IsBold bool   `json:"isBold"`
}

type Annotation struct {
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Color     string `json:"color"`
	Thickness int    `json:"thickness"`
}

type Shape struct {
	X1        int    `json:"x1"`
	Y1        int    `json:"y1"`
	X2        int    `json:"x2"`
	Y2        int    `json:"y2"`
	ShapeType string `json:"shapeType"`
	Color     string `json:"color"`
	Thickness int    `json:"thickness"`
}

type Image struct {
	Url    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
}

type BoardEvent struct {
	EventType  string     `json:"eventType"`
	Text       Text       `json:"text"`
	Annotation Annotation `json:"annotation"`
	Shape      Shape      `json:"shape"`
	Image      Image      `json:"image"`
	Chat       Chat       `json:"chat"`
}

const (
	ChatEventType         = "chat"
	PreviousChatEventType = "previousChat"
	BoardEventType
)

func (e *Class) BoardEventHandler() {
	defer e.LearnersLock.RUnlock()
	for event := range Classes[e.ClassId].Events {
		if event.EventType == ChatEventType || event.EventType == PreviousChatEventType {
			return
		}
		fmt.Println("WRITING EVENTS to STUDENTs")
		e.LearnersLock.RLock()
		for index, learner := range e.Learners {
			fmt.Println("FOUND LEARNERS ", index, event.EventType)
			err := learner.conn.WriteJSON(event)
			if err != nil {
				fmt.Println("Error writing event")
				return
			}
		}
	}
}
