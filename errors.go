package digigo // import "go.mozilla.org/digigo"
import "fmt"

// Errors is a list of error returned by the Digicert API
// https://www.digicert.com/services/v2/documentation/errors
type Errors struct {
	Errors []Error `json:"errors"`
}

// Error returned by Digicert has a code and a message
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Print several errors into a single string
func (errors Errors) String() string {
	if len(errors.Errors) == 1 {
		return fmt.Sprintf("%s %s",
			errors.Errors[0].Code, errors.Errors[0].Message)
	}
	var str string
	for i, e := range errors.Errors {
		if str != "" {
			str += " "
		}
		str += fmt.Sprintf("%d) %s %s", i, e.Code, e.Message)
	}
	return str
}
