package createPlaylist

import (
	"context"
	"fmt"
	"log"

	"github.com/zmb3/spotify/v2"
)

// create playlist matching target rythm and favorite artists
func Create(client *spotify.Client) {
	// wait for auth to complete
	seeds := getSeeds(client)

	createPlaylist(client, seeds)
}

// get seeds from favorite artists
func getSeeds(client *spotify.Client) spotify.Seeds {
	topArtists, topArtistsError := client.CurrentUsersTopArtists(context.Background())
	if topArtistsError != nil {
		log.Fatal(topArtistsError)
	}
	// jTopArtists, _ := json.MarshalIndent(*topArtists, "", "\t")
	// fmt.Println(string(jTopArtists))

	var seeds spotify.Seeds
	for i, artist := range topArtists.Artists {
		seeds.Artists = append(seeds.Tracks, artist.ID)
		if i >= 5 {
			break
		}
	}

	return seeds
}

// create an empty playlist and add recommandations from the seed
func createPlaylist(client *spotify.Client, seeds spotify.Seeds) {
	ctx := context.Background()

	const targetCadence = 180
	const genericPlaylistDescription = "A playlist created by SpotifyCadence"

	user, err := client.CurrentUser(ctx)
	if err != nil {
		log.Fatal(err)
	}

	trackAttributes := spotify.NewTrackAttributes().
		MinTempo(targetCadence/2 - 2).
		MaxTempo(targetCadence/2 + 3)
	reco, recoError := client.GetRecommendations(ctx, seeds, trackAttributes, spotify.Country(user.Country), spotify.Limit(50))
	if recoError != nil {
		log.Fatal("reco error:", recoError)
	}

	fmt.Println("size reco :", len(reco.Tracks))
	// jreco, _ := json.MarshalIndent(reco.Tracks, "", "\t")
	// fmt.Printf(string(jreco))

	emptyPlaylist, playlistCreationErr := client.CreatePlaylistForUser(ctx, user.ID, fmt.Sprintf("spotifyCadence: %d ppm", targetCadence), genericPlaylistDescription, true, false)
	if playlistCreationErr != nil {
		log.Fatal("playlist creation error:", playlistCreationErr)
	}

	recoID := make([]spotify.ID, 0, len(reco.Tracks))
	for _, v := range reco.Tracks {
		recoID = append(recoID, v.ID)
	}

	_, addTracksErr := client.AddTracksToPlaylist(ctx, emptyPlaylist.ID, recoID...)
	if addTracksErr != nil {
		log.Fatal("error adding tracks to playlist: ", addTracksErr)
	}

	fmt.Println("Successfully created playlist ", emptyPlaylist.Name)
}
