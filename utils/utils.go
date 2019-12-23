package utils

import (
    "io/ioutil"
    "os"
    "path/filepath"
    "fmt"
    "strings"
    "regexp"
    //"errors"
    "encoding/json"
    "github.com/nlopes/slack"
)

type Regexp struct { regexp.Regexp }

type MessageFilter struct {
    AuthorId string `json:"author"`
    TextPattern *Regexp `json:"text_pattern"`
}

type Source struct {
    //all
    Token string `json:"token"`
    ChannelId string `json:"channel_id"`
    // read
    Filter *MessageFilter `json:"matching"`
    ReplyFilter *MessageFilter `json:"not_replied_by"`
    // post / update
    SlackUserMap map[string]string `json:"slack_user_map"`
    // search
    Query string `json:"query"`
}

type Version map[string]string

type Metadata []MetadataField

type MetadataField struct {
    Name  string `json:"name"`
    Value string `json:"value"`
}

type InRequest struct {
    Source  Source  `json:"source"`
    Version Version `json:"version"`
    Params InParams `json:"params"`
}

type InResponse struct {
    Version  Version  `json:"version"`
    Metadata Metadata `json:"metadata"`
}

type InParams struct {
    TextPattern *Regexp `json:"text_pattern"`
    Query string `json:"query"`
}

type OutParams struct {
    Message *OutMessage `json:"message"`
    MessageFile string `json:"message_file"`
    Ts string `json:"update_ts"`
}

type OutRequest struct {
    Source  Source  `json:"source"`
    Params OutParams `json:"params"`
}

type OutMessage struct {
    Text string `json:"text"`
    Attachments []slack.Attachment `json:"attachments"`
    Blocks slack.Blocks `json:"blocks"`
    slack.PostMessageParameters
}

type OutResponse struct {
    Version  Version  `json:"version"`
    Metadata Metadata `json:"metadata"`
}

type CheckRequest struct {
    Source  Source  `json:"source"`
    Version Version `json:"version"`
}

type CheckResponse []Version

type SlackRequest struct {
    Contents string
}

func (r *Regexp) UnmarshalJSON(payload []byte) error {
    var pattern string
    err := json.Unmarshal(payload, &pattern)
    if err != nil { return err }

    regexp, regexp_err := regexp.Compile(pattern)
    if regexp_err != nil { return regexp_err }

    *r = Regexp{*regexp}

    return nil
}

func Interpolate(text string, source_dir string, source *Source) string {

    var out_text string

    start_var := 0
    end_var := 0
    inside_var := false
    c0 := '_'

    for pos, c1 := range text {
        if inside_var {
            if c0 == '}' && c1 == '}' {
                inside_var = false
                end_var = pos + 1

                var value string
                var var_name_proc []string

                if text[start_var+2] == '$' {
                    var_name := text[start_var+3:end_var-2]
                    var_name_proc = strings.Split(var_name, "|")
                    var_name = var_name_proc[0]
                    value = os.Getenv(var_name)
                    fmt.Fprintf(os.Stderr, "- Interpolating "+ var_name +"\n")
                } else {
                    var_name := text[start_var+2:end_var-2]
                    var_name_proc = strings.Split(var_name, "|")
                    var_name = var_name_proc[0]
                    value = Get_file_contents(filepath.Join(source_dir, var_name))
                    fmt.Fprintf(os.Stderr, "- Interpolating "+ var_name +"\n")
                }

                if len(var_name_proc) > 1{
                    if var_name_proc[1] == "blame" {
                        fmt.Fprintf(os.Stderr, "About to apply blame:\n")
                        fmt.Fprintf(os.Stderr, value)
                        fmt.Fprintf(os.Stderr, "\n")
                        m, _ := json.MarshalIndent(source.SlackUserMap, "", "  ")
                        fmt.Fprintf(os.Stderr, "%s\n", m)
                        fmt.Fprintf(os.Stderr, "\n")
                        value = source.SlackUserMap[value]
                    }
                }

                out_text += value
            }
        } else {
            if c0 == '{' && c1 == '{' {
                inside_var = true
                start_var = pos - 1
                out_text += text[end_var:start_var]
            }
        }
        c0 = c1
    }

    out_text += text[end_var:]

    return out_text
}

func Get_file_contents(path string) string {

    matched, err := Glob(path)
    if open_err != nil {
        err("Gloing Pattern", open_err)
    }

    path = matched[0]

    file, open_err := os.Open(path)
    if open_err != nil {
        Fatal("opening file", open_err)
    }

    data, read_err := ioutil.ReadAll(file)
    if read_err != nil {
        Fatal("reading file", read_err)
    }

    text := string(data)
    text = strings.TrimSuffix(text, "\n")

    // clean the string from \n in last possition
        
    return text
}

func Fatal(doing string, err error) {
    fmt.Fprintf(os.Stderr, "Error " + doing + ": " + err.Error() + "\n")
    os.Exit(1)
}

func Fatal1(reason string) {
    fmt.Fprintf(os.Stderr, reason + "\n")
    os.Exit(1)
}