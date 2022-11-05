# SpotifyCadence

This package is a script that creates a spotify playlist for a target running cadence (ex: 180 ppm) with recommendattions from your favorite artists

This is a quick personnal project to test golang so forgive the potentially numerous mistakes

[example playlist created](https://open.spotify.com/playlist/2vQ4zmxE7tsnite80KiUcm?si=730df7eeef064f2f)

## How to start

In order to make it work you need to:
1. Register an application at: https://developer.spotify.com/my-applications/: Use "http://localhost:8080/callback" as the redirect URI
2. Set the SPOTIFY_ID environment variable to the client ID you got in step 1.
3. Set the SPOTIFY_SECRET environment variable to the client secret from step 1.
4. Start the app with `go run SpotifyCadence.go`
5. click on the link and authorize the app (authorization list is available in the code in main package in var "auth")
6. now `go run` for real!