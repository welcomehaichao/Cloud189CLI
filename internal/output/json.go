package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type OutputFormat string

const (
	FormatJSON  OutputFormat = "json"
	FormatYAML  OutputFormat = "yaml"
	FormatTable OutputFormat = "table"
)

type Output struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

type ErrorInfo struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func NewOutput(success bool, data interface{}) *Output {
	return &Output{
		Success: success,
		Data:    data,
	}
}

func NewErrorOutput(code, message string) *Output {
	return &Output{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
	}
}

func NewErrorOutputWithDetails(code, message string, details map[string]interface{}) *Output {
	return &Output{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

func PrintJSON(out *Output) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(out)
}

func PrintJSONToWriter(w io.Writer, out *Output) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(out)
}

func PrintYAML(out *Output) error {
	return PrintYAMLToWriter(os.Stdout, out)
}

func PrintYAMLToWriter(w io.Writer, out *Output) error {
	data, err := json.Marshal(out)
	if err != nil {
		return err
	}

	var yamlData interface{}
	if err := json.Unmarshal(data, &yamlData); err != nil {
		return err
	}

	yamlStr, err := toYAML(yamlData, 0)
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(w, yamlStr)
	return err
}

func toYAML(data interface{}, indent int) (string, error) {
	switch v := data.(type) {
	case map[string]interface{}:
		result := ""
		for key, val := range v {
			result += fmt.Sprintf("%s%s:", getIndent(indent), key)
			switch val.(type) {
			case map[string]interface{}, []interface{}:
				result += "\n"
				subYAML, err := toYAML(val, indent+2)
				if err != nil {
					return "", err
				}
				result += subYAML
			default:
				result += fmt.Sprintf(" %v\n", formatValue(val))
			}
		}
		return result, nil
	case []interface{}:
		result := ""
		for _, item := range v {
			result += fmt.Sprintf("%s-", getIndent(indent))
			switch item.(type) {
			case map[string]interface{}, []interface{}:
				result += "\n"
				subYAML, err := toYAML(item, indent+2)
				if err != nil {
					return "", err
				}
				result += subYAML
			default:
				result += fmt.Sprintf(" %v\n", formatValue(item))
			}
		}
		return result, nil
	default:
		return fmt.Sprintf("%v\n", formatValue(data)), nil
	}
}

func getIndent(level int) string {
	result := ""
	for i := 0; i < level; i++ {
		result += " "
	}
	return result
}

func formatValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case bool:
		return fmt.Sprintf("%t", val)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", val)
	case float32, float64:
		return fmt.Sprintf("%v", val)
	case nil:
		return "null"
	default:
		return fmt.Sprintf("%v", val)
	}
}
