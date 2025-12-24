package connect

import (
	"time"
)

type Request struct {
	URL    string `json:"url"`
	Params string `json:"params"`
}

type InteractionLog struct {
	Gateway   string    `json:"gateway"`
	Request   *Request  `json:"request,omitempty"`
	Status    *int      `json:"status,omitempty"`
	Response  *string   `json:"response,omitempty"`
	Kind      string    `json:"kind"`
	CreatedAt time.Time `json:"created_at"` // RFC3339 by default in Go
	Duration  float64   `json:"duration"`
}

type LogWriter struct {
	created        time.Time
	kind           string
	responseStatus *int
	request        *Request
	response       *string
}

func newLogWriter(kind string) LogWriter {
	return LogWriter{
		kind:    kind,
		created: time.Now(),
	}
}

func (self *LogWriter) SetStatus(status int) {
	self.responseStatus = &status
}

func (self *LogWriter) SetResponse(response string) {
	self.response = &response
}

func (self *LogWriter) SetRequest(request string, url string) {
	self.request = &Request{URL: url, Params: request}
}

func (self LogWriter) IntoInteractionLog() InteractionLog {
	return InteractionLog{
		Gateway:   "stbl",
		Request:   self.request,
		Status:    self.responseStatus,
		Response:  self.response,
		Kind:      self.kind,
		CreatedAt: self.created,
		Duration:  time.Since(self.created).Seconds(),
	}
}

type InteractionLogs struct {
	logs    []InteractionLog
	Current *LogWriter
}

func EmptyInteractionLogs() InteractionLogs {
	return InteractionLogs{
		logs:    []InteractionLog{},
		Current: nil,
	}
}

func (self *InteractionLogs) AddLog(log LogWriter) {
	self.logs = append(self.logs, log.IntoInteractionLog())
}

// Enter new interaction log span
func (self *InteractionLogs) Enter(kind string) *LogWriter {
	if self.Current != nil {
		self.logs = append(self.logs, self.Current.IntoInteractionLog())
	}
	newWriter := newLogWriter(kind)
	self.Current = &newWriter
	return &newWriter
}

// This method should be called once
func (self InteractionLogs) IntoInner() []InteractionLog {
	if self.Current != nil {
		self.logs = append(self.logs, self.Current.IntoInteractionLog())
	}
	return self.logs
}
