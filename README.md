# mars-rover-twitter-bot
Mars Rover Photos Twitter bot

## Installing it
```shell
$ go get github.com/odeke-em/mars-rover-twitter-bot
$ mars-rover-twitter-bot --help # For help
$ mars-rover-twitter-bot --offset-in-hours 72 # Start from 72 hours ago
```

## Environment variables
Name | Required | Details
---|---|---|---
MARS_ROVER_TWITTER_CONSUMER_KEY|true|Your Twitter Consumer Key. Go to https://apps.twitter.com for one
MARS_ROVER_TWITTER_CONSUMER_SECRET|true|Your Twitter Consumer Secret. Go to https://apps.twitter.com for one
MARS_ROVER_TWITTER_ACCESS_TOKEN|true|Your Twitter Access Token. Go to https://apps.twitter.com for one
MARS_ROVER_TWITTER_ACCESS_SECRET|true|Your Twitter Access Key. Go to https://apps.twitter.com for one
