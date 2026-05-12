package incident

type IncidentPayload struct {
	Service    string `json:"service"`
	Level      string `json:"level"`
	Message    string `json:"message"`
	Type       string `json:"type"`
	Timestamp  string `json:"timestamp"`
	TraceId    string `json:"trace_id"`
	StackTrace string `json:"stack_trace"`
	Path       string `json:"path"`
	Method     string `json:"method"`
	StatusCode int    `json:"status_code"`
}

func (i IncidentPayload) GetService() string    { return i.Service }
func (i IncidentPayload) GetLevel() string      { return i.Level }
func (i IncidentPayload) GetMessage() string    { return i.Message }
func (i IncidentPayload) GetType() string       { return i.Type }
func (i IncidentPayload) GetTimestamp() string  { return i.Timestamp }
func (i IncidentPayload) GetTraceId() string    { return i.TraceId }
func (i IncidentPayload) GetStackTrace() string { return i.StackTrace }
func (i IncidentPayload) GetPath() string       { return i.Path }
func (i IncidentPayload) GetMethod() string     { return i.Method }
func (i IncidentPayload) GetStatusCode() int    { return i.StatusCode }
