/*
This package is a script that creates a spotify playlist for a target running cadence (ex: 180 ppm) with recommendattions from your favorite artists

In order to make it work you need to:
1. Register an application at: https://developer.spotify.com/my-applications/: Use "http://localhost:8080/callback" as the redirect URI
2. Set the SPOTIFY_ID environment variable to the client ID you got in step 1.
3. Set the SPOTIFY_SECRET environment variable to the client secret from step 1.

TODO:
- make some arguments available for:
  - number of artists,
  - target size for playlist

- manual deploy on a small server
*/
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	spotifyauth "github.com/zmb3/spotify/v2/auth"

	"github.com/zmb3/spotify/v2"

	"SpotifyCadence/createPlaylist"

	"github.com/joho/godotenv"
)

/*
1. get an auth token using a tmp server and a callback
2. get favorite artists
3. create  playlist matching target rythm and favorite artists
*/
func main() {
	err := godotenv.Load("env/secrets.env")
	if err != nil {
		log.Fatalf("error loading env files")
	}

	cadence := flag.Float64("cadence", 180, "target cadence or running pace, most run around a 150")
	flag.Parse()

	// start authentication prompt and wait for channel
	getSpotifyClient()
	client := <-ch

	createPlaylist.Create(client, *cadence)
}

// AUTHENTICATION

// redirectURI is the OAuth redirect URI for the prompt.
const redirectURI = "http://localhost:8080/callback"

var (
	auth = spotifyauth.New(
		spotifyauth.WithRedirectURL(redirectURI),
		spotifyauth.WithScopes(spotifyauth.ScopeUserReadPrivate, spotifyauth.ScopeUserTopRead, spotifyauth.ScopePlaylistModifyPublic),
	)
	ch    = make(chan *spotify.Client)
	state = "abc123"
)

/*
create a tmp http server via a goroutine to authenticate via callback
main function waits for authentication to be completed via the [ch] channel in [completeAuth] function
*/
func getSpotifyClient() {
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
}

// call back function for auth with channel bridge
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
