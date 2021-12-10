package main

import (
    "os"
	"fmt"
	"strings"
	"log"
    "context"
    "math/rand"
    "time"
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"
    "github.com/jasonlvhit/gocron"
    "github.com/zmb3/spotify/v2"
    "github.com/zmb3/spotify/v2/auth"
    "golang.org/x/oauth2/clientcredentials"
    twilio "github.com/twilio/twilio-go"
    openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

func main() {

    err := godotenv.Load(".env")

    if err != nil {
        log.Fatalf("Error loading .env file")
    }
	router := gin.Default()
	router.GET("/sms", smsEndpoint)
	router.Run(":5000")
}

func smsEndpoint(c *gin.Context) {
    songOfTheDayIntro := ""
    cookie := "NotSet"
    cookie, err := c.Cookie("song_cookie")

    if err != nil {
        songOfTheDayIntro = "Welcome to ʕ•́ᴥ•̀ʔっ🎵 SONG OF THE DAY 🎵 Would you like to get a new featured song sent to you everyday? Respond with 🎵 to sign up or respond with anything else to get sent a new track. You can also opt out at any time."
        c.SetCookie("song_cookie", "test", 3600, "/sms", "localhost", false, true)
        fmt.Printf("Just set cookie value")
        c.Header("Content-Type", "application/xml")
        c.String(http.StatusOK, "<Response><Message>" + songOfTheDayIntro + "</Message></Response>")
        return
    }

    receivedMsg := c.Request.URL.Query()["Body"][0]
    toPhone := c.Request.URL.Query()["From"][0]
    fmt.Printf("Cookie value: %s \n", cookie)

    if receivedMsg == "🎵" {
        go func() {
            gocron.Every(1).Day().At("10:30").Do(sendSong, toPhone)
            <- gocron.Start()
        }()
        c.Header("Content-Type", "application/xml")
        c.String(http.StatusOK, "<Response><Message>Thanks for signing up. You'll get your new song of the day every morning. Reply with 🛑 to opt-out at any time or respond with anything else to get a song right now.</Message></Response>")
    } else if receivedMsg == "🛑" {
        gocron.Remove(sendSong)
        gocron.Clear()
        c.Header("Content-Type", "application/xml")
        c.String(http.StatusOK, "<Response><Message>Sorry to see you go! Just respond with 🎵 to sign up again or message here anytime to get a new song on-demand.</Message></Response>")
    } else {

        msgs := []string{"In a dancing mood?", "Check this one out:", "You'll love this one:", "Want to get inspired?",
        "Here ya go!", "Another new song:", "Jam out to this!", "Here's a new jam for you!", "Ta da!!!", "Take a listen:",
        "Hope you'll like this one:", "Here's a featured song for you:", "Just for you!!", "Yay for new music!!",}
        rand.Seed(time.Now().UnixNano())

        numMsgs := len(msgs)
        newSong := getRandomSong()
    	c.Header("Content-Type", "application/xml")
    	c.String(http.StatusOK, "<Response><Message>" + msgs[rand.Intn(numMsgs)] + "</Message><Message>" + newSong + "</Message></Response>")
    }
}

func getRandomSong() string {

    ctx := context.Background()
    config := &clientcredentials.Config{
        ClientID: os.Getenv("SPOTIFY_ID"),
        ClientSecret: os.Getenv("SPOTIFY_SECRET"),
        TokenURL: spotifyauth.TokenURL,
    }
    token, err := config.Token(ctx)
    if err != nil {
        log.Fatalf("couldn't get token: %v", err)
    }

    httpClient := spotifyauth.New().Client(ctx, token)
    client := spotify.New(httpClient)
    _, page, err := client.FeaturedPlaylists(ctx)
    if err != nil {
        log.Fatalf("couldn't get featured playlists: %v", err)
    }

    rand.Seed(time.Now().UnixNano())

    numPlaylists := len(page.Playlists)
    randomPlaylist := page.Playlists[rand.Intn(numPlaylists)].ID

    tracks, err := client.GetPlaylistTracks(ctx, randomPlaylist)
    if err != nil {
         log.Fatalf("couldn't get playlist tracks: %v", err)
     }

    numTracks := len(tracks.Tracks)
    randomTrack := string(tracks.Tracks[rand.Intn(numTracks)].Track.URI)
    trackID := strings.Split(randomTrack, ":")[2]

    trackURL := "https://open.spotify.com/track/" + trackID

    return string(trackURL)
}

func sendSong(toPhone string) {

    newSong := getRandomSong()
    accountSid := os.Getenv("TWILIO_ACCOUNT_SID")
    authToken := os.Getenv("TWILIO_AUTH_TOKEN")

    twilioClient := twilio.NewRestClientWithParams(twilio.RestClientParams{
        Username: accountSid,
        Password: authToken,
    })

    params := &openapi.CreateMessageParams{}
    params.SetTo(toPhone)
    params.SetFrom(os.Getenv("FROM_PHONE"))
    params.SetBody(newSong)

    resp, err := twilioClient.ApiV2010.CreateMessage(params)
    if err != nil {
        fmt.Println(err.Error())
        err = nil
    } else {
        fmt.Println("Message Sid: " + *resp.Sid)
    }
}
