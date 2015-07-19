package css

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"path"
	"regexp"
	"strings"
)

func minifyVal(value string, file string) string {
	value = cleanSpaces(value)

	// Values need special care
	value = cleanHex(value)

	if (len(file) > 0) {
		value = cleanUrl(value, file)
	}

	return value
}

func cleanHex(value string) string {
	re := regexp.MustCompile(`#(\d{6})`)
	matches := re.FindAllString(value, -1)
	for _, hex := range matches {
		if isFull(hex) {
			r := strings.NewReplacer(hex, newHex(hex))
			value = r.Replace(value)
		}
	}
	return value
}

func isFull(hex string) bool {
	cmp := []byte(hex)[1:]
	for _, l := range cmp {
		if cmp[0] != l {
			return false
		}
	}
	return true
}

func newHex(hex string) string {
	return string([]byte(hex)[:4])
}

func cleanUrl(value string, file string) string {
	re := regexp.MustCompile(`url\((.*)\)`)
	matches := re.FindStringSubmatch(value)
	if len(matches) != 2 {
		return value
	}

	img := removeQuotes(matches[1])
	if string([]byte(img)[0:4]) == "http" {
		img = getWebImg(img)
	} else {
		img = getLocalImg(img, file)
	}
	return img
}

/**
 * If there are quotes or double quotes around the url, take them off.
 */
func removeQuotes(img string) string {
	bytedImg := []byte(img)
	first := bytedImg[0]
	if first == '\'' || first == '"' {
		return string(bytedImg[1 : len(bytedImg)-1])
	}
	return img
}

func getWebImg(img string) string {
	output, err := http.Get(img)
	if err != nil {
		panic(err)
	}
	defer output.Body.Close()

	body, err := ioutil.ReadAll(output.Body)
	if err != nil {
		panic(err)
	}
	return writeWebUrl(body, output)
}

func getLocalImg(img string, file string) string {
	body, err := ioutil.ReadFile(path.Dir(file) + "/" + img)
	if err != nil {
		panic(err)
	}
	return writeUrl(body, mime.TypeByExtension(img))
}

func base64Encode(data []byte) string {
	var buf bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	encoder.Write(data)
	encoder.Close()
	return buf.String()
}

func writeWebUrl(body []byte, output *http.Response) string {
	return writeUrl(body, output.Header.Get("Content-Type"))
}

func writeUrl(body []byte, mime string) string {
	var buf bytes.Buffer
	buf.WriteString("url(data:")
	buf.WriteString(mime)
	buf.WriteString(";base64,")
	buf.WriteString(base64Encode(body))
	buf.WriteString(")")
	return buf.String()
}

func stripLetter(content []byte) (byte, []byte) {
	var letter byte
	if len(content) != 0 {
		letter = content[0]
		content = content[1:]
	} else {
		content = []byte{}
	}
	return letter, content
}

func readFile(root string) string {
	content, err := ioutil.ReadFile(root)
	if err != nil {
		panic(err)
	}
	return string(content)
}

func showSelectors(selector string) (output string) {
	selectors := strings.Split(selector, ",")
	for i, sel := range selectors {
		output += minifySelector(sel)
		if i != len(selectors)-1 {
			output += ","
		}
	}
	return
}

func minifySelector(sel string) string {
	return cleanSpaces(sel)
}

func showPropVals(pairs []Pair, file string) (output string) {
	for i, pair := range pairs {
		output += fmt.Sprintf("%s:%s", minifyProp(string(pair.property)), minifyVal(string(pair.value), file))

		// Let's gain some space: semicolons are optional for the last value
		if i != len(pairs)-1 {
			output += ";"
		}
	}
	return
}

func minifyProp(property string) string {
	return cleanSpaces(property)
}

func cleanSpaces(str string) string {
	return spaceRegexp.ReplaceAllString(strings.TrimSpace(str), " ")
}
