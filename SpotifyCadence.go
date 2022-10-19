// This example demonstrates how to authenticate with Spotify using the authorization code flow.
// In order to run this example yourself, you'll need to:
//
//  1. Register an application at: https://developer.spotify.com/my-applications/
//     - Use "http://localhost:8080/callback" as the redirect URI
//  2. Set the SPOTIFY_ID environment variable to the client ID you got in step 1.
//  3. Set the SPOTIFY_SECRET environment variable to the client secret from step 1.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	spotifyauth "github.com/zmb3/spotify/v2/auth"

	"github.com/zmb3/spotify/v2"

	"github.com/joho/godotenv"
)

// redirectURI is the OAuth redirect URI for the application.
// You must register an application at Spotify's developer portal
// and enter this value.
const redirectURI = "http://localhost:8080/callback"

var (
	auth = spotifyauth.New(
		spotifyauth.WithRedirectURL(redirectURI),
		spotifyauth.WithScopes(spotifyauth.ScopeUserReadPrivate, spotifyauth.ScopeUserTopRead, spotifyauth.ScopePlaylistModifyPublic),
	)
	ch    = make(chan *spotify.Client)
	state = "abc123"
)

func main() {
	err := godotenv.Load("env/secrets.env")
	if err != nil {
		log.Fatalf("error loading env files")
	}

	// create a tmp goroutine http server to authenticate via callback
	http.HandleFunc("/callback", completeAuth)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
	})
	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatal(err)
		}
	}()

	url := auth.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)

	// wait for auth to complete
	client := <-ch
	ctx := context.Background()

	// use the client to make calls that require authorization
	user, err := client.CurrentUser(ctx)
	if err != nil {
		log.Fatal(err)
	}

	topArtists, topArtistsError := client.CurrentUsersTopArtists(ctx)
	if topArtistsError != nil {
		fmt.Println(topArtistsError)
		os.Exit(1)
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

	const targetCadence = 180
	const genericPlaylistDescription = "A playlist created by SpotifyCadence"

	trackAttributes := spotify.NewTrackAttributes().
		MinTempo(targetCadence/2 - 2).
		MaxTempo(targetCadence/2 + 3)
	reco, recoError := client.GetRecommendations(ctx, seeds, trackAttributes, spotify.Country(user.Country), spotify.Limit(50)) //TODO make a script parameter limit
	if recoError != nil {
		fmt.Println("reco error:", recoError)
		os.Exit(1)
	}

	fmt.Println("size reco :", len(reco.Tracks))
	// jreco, _ := json.MarshalIndent(reco.Tracks, "", "\t")
	// fmt.Printf(string(jreco))

	emptyPlaylist, playlistCreationErr := client.CreatePlaylistForUser(ctx, user.ID, fmt.Sprintf("spotifyCadence: %d ppm", targetCadence), genericPlaylistDescription, true, false)
	if playlistCreationErr != nil {
		fmt.Println("playlist creation error:", playlistCreationErr)
		os.Exit(1)
	}

	recoID := make([]spotify.ID, 0, len(reco.Tracks))
	for _, v := range reco.Tracks {
		recoID = append(recoID, v.ID)
	}

	_, addTracksErr := client.AddTracksToPlaylist(ctx, emptyPlaylist.ID, recoID...)
	if addTracksErr != nil {
		fmt.Println("error adding tracks to playlist: ", addTracksErr)
		os.Exit(1)
	}

	fmt.Println("Successfully created playlist ", emptyPlaylist.Name)
	os.Exit(0)
}

func completeAuth(w http.ResponseWriter, r *http.Request) {
	tok, err := auth.Token(r.Context(), state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}

	// use the token to get an authenticated client
	client := spotify.New(auth.Client(r.Context(), tok))
	fmt.Fprintf(w, "Login Completed!")
	ch <- client
}
