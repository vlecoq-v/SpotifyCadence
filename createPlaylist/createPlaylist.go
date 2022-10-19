package createPlaylist

import (
	"context"
	"fmt"
	"log"

	"github.com/zmb3/spotify/v2"
)

// create playlist matching target rythm and favorite artists
// create an empty playlist and add recommandations from the seed
func Create(client *spotify.Client) {
	ctx := context.Background()
	const targetCadence = 180
	const genericPlaylistDescription = "A playlist created by SpotifyCadence"

	user, err := client.CurrentUser(ctx)
	if err != nil {
		log.Fatal(err)
	}

	seeds := getSeeds(client)
	tracks := getTracks(client, seeds, user, targetCadence)

	emptyPlaylist, playlistCreationErr := client.CreatePlaylistForUser(ctx, user.ID, fmt.Sprintf("spotifyCadence: %d ppm", targetCadence), genericPlaylistDescription, true, false)
	if playlistCreationErr != nil {
		log.Fatal("playlist creation error:", playlistCreationErr)
	}

	recoID := make([]spotify.ID, 0, len(tracks))
	for _, v := range tracks {
		recoID = append(recoID, v.ID)
	}

	for i := 0; i < len(tracks); i += 100 {
		recoIdSubSlice := recoID[i:min(i+100, cap(recoID))]
		_, addTracksErr := client.AddTracksToPlaylist(ctx, emptyPlaylist.ID, recoIdSubSlice...)
		if addTracksErr != nil {
			log.Fatal("error adding tracks to playlist: ", addTracksErr)
		}
	}

	fmt.Println("Successfully created playlist ", emptyPlaylist.Name)
}

// get Tracks from seeds
func getTracks(client *spotify.Client, seeds spotify.Seeds, user *spotify.PrivateUser, targetCadence float64) []spotify.SimpleTrack {
	trackAttributes := spotify.NewTrackAttributes().
		MinTempo(targetCadence/2 - 2).
		MaxTempo(targetCadence/2 + 3).
		MinEnergy(0.8)

	var tracks []spotify.SimpleTrack

	ctx := context.Background()
	for i := 0; i < len(seeds.Artists); i += 5 {
		seedSubSlice := spotify.Seeds{
			Artists: seeds.Artists[i:min(i+5, cap(seeds.Artists))],
		}
		recommendations, recoError := client.GetRecommendations(ctx, seedSubSlice, trackAttributes, spotify.Country(user.Country), spotify.Limit(50))
		if recoError != nil {
			log.Fatal("reco error:", recoError)
		}
		tracks = append(tracks, recommendations.Tracks...)
		fmt.Println("size reco :", len(recommendations.Tracks))
	}

	fmt.Println("total size reco :", len(tracks))
	// jreco, _ := json.MarshalIndent(reco.Tracks, "", "\t")
	// fmt.Printf(string(jreco))

	return tracks
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
	var artistNames []string
	for _, artist := range topArtists.Artists[:10] {
		seeds.Artists = append(seeds.Artists, artist.ID)
		artistNames = append(artistNames, artist.Name)
	}
	fmt.Println("size of seeds: ", len(seeds.Artists), "artist used", artistNames)

	return seeds
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
