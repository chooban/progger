package server

// ClientSettingDto represents a single key-value setting
type ClientSettingDto struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ClientSettingsResponse represents the response format (map of key-value pairs)
// This is a type alias to clarify usage
type ClientSettingsResponse map[string]interface{}
