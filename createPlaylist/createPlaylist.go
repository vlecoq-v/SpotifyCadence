package createPlaylist

import (
	"context"
	"fmt"
	"log"

	"github.com/zmb3/spotify/v2"
)

// create playlist matching target rythm and favorite artists
// create an empty playlist and add recommandations from the seed
func Create(client *spotify.Client, cadence float64) {
	ctx := context.Background()
	const genericPlaylistDescription = "A playlist created by SpotifyCadence"

	user, err := client.CurrentUser(ctx)
	if err != nil {
		log.Fatal(err)
	}

	seeds := getSeeds(client)
	tracks := getTracks(client, seeds, user, cadence)

	emptyPlaylist, playlistCreationErr := client.CreatePlaylistForUser(ctx, user.ID, fmt.Sprintf("spotifyCadence: %.0f cadence", cadence), genericPlaylistDescription, true, false)
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

	return tracks
}

// get seeds from favorite artists
// take 10 top artists from each existing spotify time ranges
func getSeeds(client *spotify.Client) spotify.Seeds {
	var seeds spotify.Seeds
	var artistNames []string
	timeRanges := []spotify.Range{
		spotify.ShortTermRange,
		spotify.MediumTermRange,
		spotify.LongTermRange,
	}

	for _, timeRange := range timeRanges {
		topArtists, topArtistsError := client.CurrentUsersTopArtists(context.Background(), spotify.Timerange(timeRange))
		if topArtistsError != nil {
			log.Fatal(topArtistsError)
		}

		for _, artist := range topArtists.Artists[:10] {
			seeds.Artists = append(seeds.Artists, artist.ID)
			artistNames = append(artistNames, artist.Name)
		}
	}
	seeds.Artists = removeDuplicateID(seeds.Artists)
	artistNames = removeDuplicateString(artistNames)

	fmt.Println("size of seeds: ", len(seeds.Artists), "artist used", artistNames)

	return seeds
}
