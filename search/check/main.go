package main

import (
    "encoding/json"
    //"io"
    //"ioutil"
    "os"
    //"os/exec"
    "fmt"
    //"strings"
    //"net/http"
    "github.com/jleben/slack-chat-resource/utils"
    "github.com/nlopes/slack"
)

func main() {

    var request utils.CheckRequest

    var err error

	err = json.NewDecoder(os.Stdin).Decode(&request)
	if err != nil {
		fatal("Parsing request.", err)
	}

	if len(request.Source.Token) == 0 {
        fatal1("Missing source field: token.")
    }

    if len(request.Source.ChannelId) == 0 {
        fatal1("Missing source field: channel_id.")
    }

    if len(request.Source.Query) == 0 {
        fatal1("Missing source field: query.")
    }

    slack_client := slack.New(request.Source.Token)

    messages := get_messages(&request, slack_client)

    versions := []utils.Version{}

    for _, msg := range messages.Matches {
        version := utils.Version{ "timestamp": msg.Timestamp }
        versions = append(versions, version)
    }

    response := utils.CheckResponse{}
    for i := len(versions) - 1; i >= 0; i--  {
        response = append(response, versions[i])
    }

    {
        err := json.NewEncoder(os.Stdout).Encode(&response)
        if err != nil {
            fatal("serializing response", err)
        }
    }
    
}

func get_messages(request *utils.CheckRequest, slack_client *slack.Client) *slack.SearchMessages {

    params := slack.NewSearchParameters()

    slack_client.SearchMessages(request.Source.Query, params)

    var messages *slack.SearchMessages
    messages, err := slack_client.SearchMessages(request.Source.Query, params)
    if err != nil {
        fatal("getting messages.", err)
    }

    return messages
}

func fatal(doing string, err error) {
    fmt.Fprintf(os.Stderr, "error " + doing + ": " + err.Error() + "\n")
	os.Exit(1)
}

func fatal1(reason string) {
    fmt.Fprintf(os.Stderr, reason + "\n")
    os.Exit(1)
}
