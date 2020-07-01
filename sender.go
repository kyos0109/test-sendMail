package main

import (
    "fmt"
    "flag"
    "log"
    "gopkg.in/gomail.v2"
    "io/ioutil"
    "os"
    "bufio"
    "time"
    "encoding/json"
    "io"
)

var (
    help bool
    sendLogPath string
)


type SenderData struct {
    Subject   string
    Message   string
    MailList  string
    ToolsArgs ToolsArgs

    SendersProfile []SenderProfile `json:"sendersprofile"`
}

type ToolsArgs struct {
    Timeout               time.Duration
    DelayToSend           time.Duration
    SendLogPath           string
    SendProfileConfigPath string
}

type SenderProfile struct {
    SMTPHost     string `json:"SMTPHost"`
    UserName     string `json:"UserName"`
    Passowrd     string `json:"Passowrd"`
    MailFrom     string `json:"MailFrom"`
    MailFromName string `json:"MailFromName"`
}

func (sd *SenderData) Init() {
    flag.BoolVar(&help, "help", false, "This Help")
    flag.StringVar(&sd.Subject, "s", "！！！ＩＭＰＯＲＴＡＮＴ！！！", "Message `Subject`")
    flag.StringVar(&sd.Message, "m", "edm.html", "Message Content `File Path`")
    flag.StringVar(&sd.MailList, "l", "mail_list", "MailTo User List `File Path`")
    flag.StringVar(&sd.ToolsArgs.SendLogPath, "log", "send.log", "Send Log `File Path`")
    flag.StringVar(&sd.ToolsArgs.SendProfileConfigPath, "c", "senders.json", "Send Profile `Config File Path`")
    flag.DurationVar(&sd.ToolsArgs.Timeout, "t", 5 * time.Second, "Send to STMP `Timeout`")
    flag.DurationVar(&sd.ToolsArgs.DelayToSend, "delay", 200 * time.Millisecond, "Send to STMP `Delay`")
    flag.Parse()

    sendLogPath = sd.ToolsArgs.SendLogPath

    if help {
        flag.Usage()
        os.Exit(1)
    }
}

func usage() {
    fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
    flag.PrintDefaults()
}

func main() {

    var sd SenderData
    sd.Init()

    dat, err := ioutil.ReadFile(sd.Message)
    if err != nil {
        log.Fatalf("***Error: %s", err)
    }
    log.Println("Open Mail Content File Successfully.")

    mailList, err := readLines(sd.MailList)
    if err != nil {
        log.Fatalf("***Error: %s", err)
    }
    log.Println("Open Mail List File Successfully.")

    sendersJson, err := ioutil.ReadFile(sd.ToolsArgs.SendProfileConfigPath)
    if err != nil {
        log.Fatalf("***Error: %s", err)
    }
    log.Println("Open Senders Config File Successfully.")

    if err := json.Unmarshal(sendersJson, &sd); err != nil {
        fmt.Println(err)
    }

    ch := make(chan string)

    for si, sp := range sd.SendersProfile {
        log.Printf("Send Starting... %s", sp.MailFrom)

        d := gomail.NewDialer(sp.SMTPHost, 587, sp.UserName, sp.Passowrd)

        for i, r := range mailList {
            time.Sleep(sd.ToolsArgs.DelayToSend)
            go func(r string, i, si int) {
                __connect, err := d.Dial()
                if err != nil {
                    panic(err)
                }

                m := buildMailContent(&sd, &r, &dat, &si)

                if err := gomail.Send(__connect, m); err != nil {
                    ch <- fmt.Sprintf("Error: Could not send email to %q: %v", &r, err)
                } else {
                    ch <- fmt.Sprintf("Num: %d, %s Send --> %q", i+1, sd.SendersProfile[si].MailFrom, r)
                }
                m.Reset()
            }(r, i, si)
        }

    }

    for {
        select {
        case r := <-ch:
            writeLog(r)

        case <-time.After(sd.ToolsArgs.Timeout):
            log.Println("--> Bye...")
            return
        }
    }
}

func buildMailContent (sd *SenderData, r *string, dat *[]byte, si *int) *gomail.Message {
    m := gomail.NewMessage(gomail.SetEncoding(gomail.Base64))
    m.SetHeader("From", m.FormatAddress(sd.SendersProfile[*si].MailFrom, sd.SendersProfile[*si].MailFromName))
    m.SetHeader("To", *r)
    m.SetHeader("Subject", sd.Subject)
    m.SetHeader("X-Spam-Flag", "YES")
    m.SetHeader("X-Spam-Level", "*************")
    m.SetHeader("X-Spam-Status", "Yes, score=13.7 required=5.0 tests=BAYES_99,HS_INDEX_PARAM,HTML_MESSAGE,RCVD_IN_PBL,RCVD_IN_SORBS_DUL,RDNS_NONE,URIBL_AB_SURBL,URIBL_BLACK,URIBL_JP_SURBL,URIBL_SBL,URIBL_WS_SURBL autolearn=spam version=3.2.5")
    m.SetBody("text/html", string(*dat))

    return m
}

func writeLog(l string) {
    f, err := os.OpenFile(sendLogPath, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
    if err != nil {
        log.Fatalf("error opening file: %v", err)
    }
    defer f.Close()

    wrt := io.MultiWriter(os.Stdout, f)
    log.SetOutput(wrt)
    log.Println(l)
}

func readLines(path string) ([]string, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var lines []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        lines = append(lines, scanner.Text())
    }
    return lines, scanner.Err()
}
