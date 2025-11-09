package fonts

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// create Axes struct to handle extra key for variable fonts
type Axes struct {
	Tag   string `json:"tag"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}

// extract font file path url from Google Fonts API JSON response
type Font struct {
	Items []struct {
		Family   string            `json:"family"`
		Variants []string          `json:"variants"`
		Files    map[string]string `json:"files"`
		Axes     []*Axes           `json:"axes,omitempty"`
	} `json:"items"`
}

func ParseFontFamily(fontFamily string) (parsedFontFamily string) {
	// convert font input to lowercase
	fontFamily = cases.Lower(language.Und).String(fontFamily)
	// convert first letter of each word to uppercase
	fontFamily = cases.Title(language.Und).String(fontFamily)
	// replace spaces with + for url formatting
	parsedFontFamily = strings.ReplaceAll(fontFamily, " ", "+")
	return parsedFontFamily
}

func GetFontUrl(key, fontFamily string) (fontResponse Font) {
	url := "https://www.googleapis.com/webfonts/v1/webfonts?key=" + fmt.Sprint(key) + "&family=" + fontFamily + "&capability=WOFF2&capability=VF"

	slog.Debug("Fetching font data from", "url", url)
	// Make the GET request
	res, err := http.Get(url)
	if err != nil {
		fmt.Println("Error: failed to create connection to remote host", err)
		os.Exit(1)
	}
	defer res.Body.Close()

	// check response and handle errors
	if res.StatusCode == 200 {
		// Read the response body
		body, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Println("Error: Could not read response body", err)
			os.Exit(1)
		}

		// parse the response body into the Font object struct
		var fontResponse Font
		err = json.Unmarshal(body, &fontResponse)
		if err != nil {
			fmt.Println("Error: could not parse json response", err)
			os.Exit(1)
		}
		return fontResponse
	} else if res.StatusCode == 400 {
		fmt.Println("Error: invalid API Key")
		os.Exit(1)
		return
	} else if res.StatusCode == 500 {
		fmt.Println("Error: could not find specified font:", fontFamily)
		os.Exit(1)
		return
	} else {
		fmt.Println("An unexpected error occured")
		os.Exit(1)
		return
	}
}
