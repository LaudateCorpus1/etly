package etly

import (
	"hash/fnv"
	"net/url"
	"path"
	"strings"
	"time"

	"io"

	"github.com/viant/toolbox"
)

const timeVariableExpr = "<dateFormat:"
const modeVarableExpr = "<mod:"

var jsonDecoderFactory = toolbox.NewJSONDecoderFactory()
var jsonEncoderFactory = toolbox.NewJSONEncoderFactory()

func expandDateExpressionIfPresent(text string, sourceTime *time.Time) string {
	for j := 0; j < len(text); j++ {
		matchingExpression, dateFormat := getTimeVariableIfPresent(text)
		if matchingExpression == "" {
			break
		}
		text = expandDateFormatExprIfPresent(text, matchingExpression, dateFormat, sourceTime)
	}
	return text
}

func expandModExpressionIfPresent(text string, hash int) string {
	for j := 0; j < len(text); j++ {
		matchingExpression, mod := getModVariableIfPresent(text)
		if matchingExpression == "" {
			break
		}

		if strings.Contains(text, matchingExpression) {
			var value = toolbox.AsString(hash % mod)
			text = strings.Replace(text, matchingExpression, value, len(text))
		}
		return text
	}
	return text
}

func expandDateFormatExprIfPresent(text, matchingExpr, dateFormat string, sourceTime *time.Time) string {
	if strings.Contains(text, matchingExpr) {
		var value = sourceTime.Format(toolbox.DateFormatToLayout(dateFormat))
		text = strings.Replace(text, matchingExpr, value, len(text))
	}
	return text
}

func getTimeVariableIfPresent(text string) (string, string) {
	timeMatchIndex := strings.Index(text, timeVariableExpr)
	if timeMatchIndex == -1 {
		return "", ""
	}
	timeMatchIndex += len(timeVariableExpr)
	var timeFormat = ""
	for j := 0; j < 10; j++ {
		var aChar = text[timeMatchIndex+j : timeMatchIndex+j+1]
		if aChar != ">" {
			timeFormat += aChar
		} else {
			break
		}
	}
	return timeVariableExpr + timeFormat + ">", timeFormat
}

func getModVariableIfPresent(text string) (string, int) {
	modMatchingIndex := strings.Index(text, modeVarableExpr)
	if modMatchingIndex == -1 {
		return "", 0
	}

	modMatchingIndex += len(modeVarableExpr)
	var mod = ""
	for j := 0; j < 10; j++ {
		if !(modMatchingIndex+j+1 < len(text)) {
			break
		}
		var aChar = text[modMatchingIndex+j : modMatchingIndex+j+1]
		if aChar != ">" {
			mod += aChar
		} else {
			break
		}
	}
	return modeVarableExpr + mod + ">", toolbox.AsInt(mod)
}

func GetCurrentWorkingDir() string {
	file, _, _ := toolbox.CallerInfo(2)
	parent, _ := path.Split(file)
	return parent
}
func expandCurrentWorkingDirectory(text string) string {
	if strings.Contains(text, "<pwd>") {
		text = strings.Replace(text, "<pwd>", GetCurrentWorkingDir(), len(text))
	}
	return text
}

func extractFileNameFromURL(URL string) string {
	parsedURL, err := url.Parse(URL)
	if err != nil {
		return ""
	}
	_, file := path.Split(parsedURL.Path)
	return file
}

func hash(text string) int {
	h := fnv.New64()
	h.Write([]byte(text))
	result := int(h.Sum64())
	if result < 0 {
		return result * -1
	}
	return result
}

func decodeJSONTarget(reader io.Reader, target interface{}) error {
	return jsonDecoderFactory.Create(reader).Decode(target)
}

func encodeJSONSource(writer io.Writer, target interface{}) error {
	return jsonEncoderFactory.Create(writer).Encode(target)
}
