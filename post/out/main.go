package main

import (
    "encoding/json"
    // "io/ioutil"
    "os"
    "path/filepath"
    "fmt"
    // "strings"
    "github.com/jleben/slack-chat-resource/utils"
    "github.com/nlopes/slack"
)

func main() {
	if len(os.Args) < 2 {
		println("usage: " + os.Args[0] + " <source>")
		os.Exit(1)
	}

    source_dir := os.Args[1]

    var request utils.OutRequest

    request_err := json.NewDecoder(os.Stdin).Decode(&request)
    if request_err != nil {
        utils.Fatal("Parsing request.", request_err)
    }

    if len(request.Source.Token) == 0 {
        utils.Fatal1("Missing source field: token.")
    }

    if len(request.Source.ChannelId) == 0 {
        utils.Fatal1("Missing source field: channel_id.")
    }

    if len(request.Params.MessageFile) == 0 && request.Params.Message == nil {
        utils.Fatal1("Missing params field: message or message_file.")
    }

    var message *utils.OutMessage

    if len(request.Params.MessageFile) != 0 {
        fmt.Fprintf(os.Stderr, "About to read this file:" + filepath.Join(source_dir,request.Params.MessageFile) + "\n")
        message = new(utils.OutMessage)
        read_message_file(filepath.Join(source_dir,request.Params.MessageFile), message)
    }else{
        message = request.Params.Message
    }
    
    {
        fmt.Fprintf(os.Stderr, "About process message (interpolation)\n")
        interpolate_message(message, source_dir, &request)
    }

    {
        fmt.Fprintf(os.Stderr, "About to send this message:\n")
        m, _ := json.MarshalIndent(message, "", "  ")
        fmt.Fprintf(os.Stderr, "%s\n", m)
    }

    slack_client := slack.New(request.Source.Token)

    var response utils.OutResponse

    if len(request.Params.Ts) == 0 {
        response = send(message, &request, slack_client)
    }else{
        request.Params.Ts = utils.Get_file_contents(filepath.Join(source_dir, request.Params.Ts))
        response = update(message, &request, slack_client)
    }

    response_err := json.NewEncoder(os.Stdout).Encode(&response)
    if response_err != nil {
        utils.Fatal("encoding response", response_err)
    }
}

func read_message_file(path string, message *utils.OutMessage) {
    file, open_err := os.Open(path)
    if open_err != nil {
        utils.Fatal("opening message file", open_err)
    }

    read_err := json.NewDecoder(file).Decode(message)
    if read_err != nil {
        utils.Fatal("reading message file", read_err)
    }
}

func interpolate_message(message *utils.OutMessage, source_dir string, request *utils.OutRequest) {
    message.Text = utils.Interpolate(message.Text, source_dir, &request.Source)
    message.ThreadTimestamp = utils.Interpolate(message.ThreadTimestamp, source_dir, &request.Source)

    for i := 0; i < len(message.Attachments); i++ {
        attachment := &message.Attachments[i]

        attachment.Fallback = utils.Interpolate(attachment.Fallback, source_dir, &request.Source)
        attachment.Title = utils.Interpolate(attachment.Title, source_dir, &request.Source)
        attachment.TitleLink = utils.Interpolate(attachment.TitleLink, source_dir, &request.Source)
        attachment.Pretext = utils.Interpolate(attachment.Pretext, source_dir, &request.Source)
        attachment.Text = utils.Interpolate(attachment.Text, source_dir, &request.Source)
        attachment.Footer = utils.Interpolate(attachment.Footer, source_dir, &request.Source)

        for j := 0; j < len(attachment.Fields); j++ {
            field := &attachment.Fields[j]
            field.Title = utils.Interpolate(field.Title, source_dir, &request.Source)
            field.Value = utils.Interpolate(field.Value, source_dir, &request.Source)
        }

        for k := 0; k < len(attachment.Actions); k++ {
            action := &attachment.Actions[k]
            action.Text = utils.Interpolate(action.Text, source_dir, &request.Source)
            action.URL = utils.Interpolate(action.URL, source_dir, &request.Source)
        }
    }

    for _, block := range message.Blocks.BlockSet {
		switch block.BlockType() {
		case slack.MBTContext:
			contextElements := block.(*slack.ContextBlock).ContextElements.Elements
			for _, elem := range contextElements {
				switch elem.MixedElementType() {
				case slack.MixedElementImage:
					// Assert the block's type to manipulate/extract values
					imageBlockElem := elem.(*slack.ImageBlockElement)
					imageBlockElem.ImageURL = utils.Interpolate(imageBlockElem.ImageURL, source_dir, &request.Source)
					imageBlockElem.AltText = utils.Interpolate(imageBlockElem.ImageURL, source_dir, &request.Source)
				case slack.MixedElementText:
					textBlockElem := elem.(*slack.TextBlockObject)
					textBlockElem.Text = utils.Interpolate(textBlockElem.Text, source_dir, &request.Source)
				}
			}
		case slack.MBTAction:
			// no interpolation
        case slack.MBTImage:
            elements :=  block.(*slack.ImageBlock)
            elements.ImageURL = utils.Interpolate(elements.ImageURL, source_dir, &request.Source)
            elements.Title.Text = utils.Interpolate(elements.Title.Text, source_dir, &request.Source)
		case slack.MBTSection:
            elements :=  block.(*slack.SectionBlock)
            elements.Text.Text = utils.Interpolate(elements.Text.Text, source_dir, &request.Source)

            for _, field := range elements.Fields {
                field.Text = utils.Interpolate(field.Text, source_dir, &request.Source)
            }

            // elements.Accessory  // no interpolation
		case slack.MBTDivider:
            // no interpolation
		}
    }
}

func update(message *utils.OutMessage, request *utils.OutRequest, slack_client *slack.Client) utils.OutResponse {

    fmt.Fprintf(os.Stderr, "About to post an update message: " + request.Params.Ts  + "\n")
    _, timestamp, _, err := slack_client.UpdateMessage(request.Source.ChannelId,
        request.Params.Ts,
        slack.MsgOptionText(message.Text, false),
        slack.MsgOptionAttachments(message.Attachments...),
        slack.MsgOptionBlocks(message.Blocks.BlockSet...),
        slack.MsgOptionPostMessageParameters(message.PostMessageParameters))

    if err != nil {
        utils.Fatal("sending", err)
    }

    var response utils.OutResponse
    response.Version = utils.Version { "timestamp": timestamp }
    return response
}

func send(message *utils.OutMessage, request *utils.OutRequest, slack_client *slack.Client) utils.OutResponse {

    fmt.Fprintf(os.Stderr, "About to post a new message \n")
    _, timestamp, err := slack_client.PostMessage(request.Source.ChannelId,
        slack.MsgOptionText(message.Text, false),
        slack.MsgOptionAttachments(message.Attachments...),
        slack.MsgOptionBlocks(message.Blocks.BlockSet...),
        slack.MsgOptionPostMessageParameters(message.PostMessageParameters))


    if err != nil {
        utils.Fatal("sending", err)
    }

    var response utils.OutResponse
    response.Version = utils.Version { "timestamp": timestamp }
    return response
}