package createPlaylist

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/zmb3/spotify/v2"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func printInJsonFormat(object any) {
	jsonObject, err := json.MarshalIndent(object, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf(string(jsonObject))
}

func removeDuplicateID(sliceList []spotify.ID) []spotify.ID {
	allKeys := make(map[spotify.ID]bool)
	list := []spotify.ID{}
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func removeDuplicateString(sliceList []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}
