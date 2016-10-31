package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/odeke-em/nasa"
)

var (
	roverClient *nasa.Client
	twitterAPI  *anaconda.TwitterApi
)

var (
	twitterConsumerKey    = envValueAtInit("MARS_ROVER_TWITTER_CONSUMER_KEY")
	twitterConsumerSecret = envValueAtInit("MARS_ROVER_TWITTER_CONSUMER_SECRET")
	twitterAccessToken    = envValueAtInit("MARS_ROVER_TWITTER_ACCESS_TOKEN")
	twitterAccessSecret   = envValueAtInit("MARS_ROVER_TWITTER_ACCESS_SECRET")
)

var initErrMsgList = []string{}

func envValueAtInit(key string) string {
	value := os.Getenv(key)
	if value == "" {
		initErrMsgList = append(initErrMsgList, fmt.Sprintf("%q is not defined", key))
	}
	return value
}

func exitOnError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}
}

func init() {
	var err error
	roverClient, err = nasa.New()
	if err != nil {
		exitOnError(err)
	}

	if len(initErrMsgList) > 0 {
		msg := fmt.Sprintf("Errors Encountered:\n%s", strings.Join(initErrMsgList, "\n"))
		exitOnError(fmt.Errorf("%s", msg))
	}

	anaconda.SetConsumerKey(twitterConsumerKey)
	anaconda.SetConsumerSecret(twitterConsumerSecret)
	twitterAPI = anaconda.NewTwitterApi(twitterAccessToken, twitterAccessSecret)
}

func tweetPhotos(periodInHours uint) error {
	offsetTime := time.Now().Add(-1 * time.Duration(periodInHours) * time.Hour)
	startTime := &offsetTime

	marsPhotos, err := roverClient.MarsPhotos(startTime)
	if err != nil {
		return err
	}

	photos := marsPhotos.Photos
	if len(photos) < 1 {
		return fmt.Errorf("no photos for %q", startTime)
	}

	// Distribute the photos over periodInhours
	rawHourlyDistribution := float32(periodInHours) / float32(len(photos))
	hours := int(rawHourlyDistribution)
	minutes := int(math.Ceil(60 * float64(rawHourlyDistribution-float32(hours))))
	pausePeriod := (time.Duration(hours) * time.Hour) + (time.Duration(minutes) * time.Minute)

	ticker := time.Tick(pausePeriod)
	log.Printf("Total Photos: %d PausePeriod: %s\n", len(photos), pausePeriod)
	for i := 0; i < len(photos); i++ {
		// tweet the photo
		photo := photos[i]
		log.Printf("#%d: About to tweet %#v\n", i, photo)
		tweetPhoto(photo)
		<-ticker
	}

	return nil
}

const tweetMsg = `PhotoId {{.Id}} Taken on {{.EarthDate}} by Rover {{.Rover.Name}} aka {{.Rover.ShortName}} launched on {{.Rover.LaunchDate}}`

var tweetTmpl = template.Must(template.New("marsPhotoTweet").Parse(tweetMsg))

func composeTweet(mphoto *nasa.MarsPhoto) (string, error) {
	buf := new(bytes.Buffer)
	if err := tweetTmpl.Execute(buf, mphoto); err != nil {
		return "", err
	}
	return string(buf.Bytes()), nil
}

func tweetPhoto(mphoto *nasa.MarsPhoto) error {
	if mphoto.ImageURL == "" {
		return fmt.Errorf("expecting an image URL")
	}
	statusMsg, err := composeTweet(mphoto)
	if err != nil {
		return err
	}

	base64OfImage, err := imageBase64(mphoto.ImageURL)
	if err != nil {
		return err
	}
	media, err := twitterAPI.UploadMedia(base64OfImage)
	if err != nil {
		return err
	}

	// For sending the actual tweet:
	// Thanks to http://stackoverflow.com/questions/30992532/go-anaconda-twitter-media-upload-with-tweet
	tw := url.Values{}
	tw.Set("media_ids", fmt.Sprintf("%v", media.MediaID))
	result, err := twitterAPI.PostTweet(statusMsg, tw)
	log.Printf("result: %v err: %v\n", result, err)
	return err
}

type tweet struct {
	MediaId uint64 `json:"media_ids"`
	Status  string `json:"status"`
}

func imageBase64(uri string) (string, error) {
	res, err := http.Get(uri)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return "", fmt.Errorf("%s", res.Status)
	}
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, res.Body); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func main() {
	var periodInHours uint
	flag.UintVar(&periodInHours, "period-in-hours", 24, "the period of the bot in hours")
	flag.Parse()

	for {
		tweetPhotos(periodInHours)
	}
}
