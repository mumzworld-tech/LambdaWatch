package loki

// PushRequest is the Loki push API request body
type PushRequest struct {
	Streams []Stream `json:"streams"`
}

// Stream represents a single log stream in Loki
type Stream struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

// NewPushRequest creates a new push request with the given labels and log values
func NewPushRequest(labels map[string]string, values [][]string) *PushRequest {
	return &PushRequest{
		Streams: []Stream{
			{
				Stream: labels,
				Values: values,
			},
		},
	}
}
