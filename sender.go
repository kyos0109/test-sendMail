package main

import (
    "fmt"
    "flag"
    "log"
    "io/ioutil"
    "os"
    "bufio"
    "time"
    "encoding/json"
    "sync"
    "gopkg.in/gomail.v2"
    "github.com/cheggaaa/pb/v3"
)

var (
    help bool
    sendLogPath string
    wgMax int
)


type SenderData struct {
    Subject         string
    MailList        []string
    MailListPath    string
    MailContentPath string
    MailContent     string
    ToolsArgs       ToolsArgs
    SendCount       int
    SendFailCount   int
    SendCountLock   sync.Mutex

    SendersProfile []SenderProfile `json:"SenderSprofile"`
}

type ToolsArgs struct {
    Timeout               time.Duration
    DelayToSend           time.Duration
    SendLogPath           string
    SendProfileConfigPath string
    SendProfileConfig     []byte
    ForLoopWait           time.Duration
    ForLoop               bool
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
    flag.BoolVar(&sd.ToolsArgs.ForLoop, "loop", false, "Open For Loop.")
    flag.StringVar(&sd.Subject, "s", "！！！ＩＭＰＯＲＴＡＮＴ！！！", "Message `Subject`")
    flag.StringVar(&sd.MailContentPath, "m", "edm.html", "Message Content `File Path`")
    flag.StringVar(&sd.MailListPath, "l", "mail_list", "MailTo User List `File Path`")
    flag.StringVar(&sd.ToolsArgs.SendLogPath, "log", "send.log", "Send Log `File Path`")
    flag.StringVar(&sd.ToolsArgs.SendProfileConfigPath, "c", "senders.json", "Send Profile `Config File Path`")
    flag.DurationVar(&sd.ToolsArgs.Timeout, "t", 5 * time.Second, "Send to STMP `Timeout`")
    flag.DurationVar(&sd.ToolsArgs.DelayToSend, "delay", 200 * time.Millisecond, "Send to STMP `Delay`")
    flag.DurationVar(&sd.ToolsArgs.ForLoopWait, "wait", 60 * time.Second, "For Loop Wait For Next.")
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
    sd.SendCount = 0
    sd.MailContent = sd.setMailContent()
    sd.MailList = sd.setMailList()
    sd.ToolsArgs.SendProfileConfig = sd.setSenderProfileConfig()


    if err := json.Unmarshal(sd.ToolsArgs.SendProfileConfig, &sd); err != nil {
        fmt.Println(err)
    }

    for {

        var wg sync.WaitGroup
        wgMax = len(sd.SendersProfile) * len(sd.MailList)
        wg.Add(wgMax)

        bar := pb.Full.Start(wgMax)
        log.Println("Starting...")
        for si, sp := range sd.SendersProfile {
            writeLog(fmt.Sprintf("---------------------------------------------"))
            writeLog(fmt.Sprintf("Send Starting... %s ", sp.MailFrom))
            writeLog(fmt.Sprintf("---------------------------------------------"))

            d := gomail.NewDialer(sp.SMTPHost, 587, sp.UserName, sp.Passowrd)

            for i, r := range sd.MailList {
                time.Sleep(sd.ToolsArgs.DelayToSend)
                go sd.doSend(r, i, si, d, &wg)
                bar.Increment()
            }
        }
        wg.Wait()
        bar.Finish()

        fmt.Println("Done.")
        fmt.Printf("Send Total Count: %d\n", sd.SendCount)
        fmt.Printf("Send Fail Total Count: %d\n", sd.SendFailCount)

        if sd.ToolsArgs.ForLoop == false {
            break
        }
        fmt.Printf("修但幾咧....\n")

        time.Sleep(sd.ToolsArgs.ForLoopWait * time.Second)
    }
}

func (sd *SenderData) doSend (
    mailTo string,
    mailToCount,
    senderConfigCount int,
    newDial *gomail.Dialer,
    wg *sync.WaitGroup) {

    defer wg.Done()
    __connect, err := newDial.Dial()
    if err != nil {
        panic(err)
    }

    m := sd.buildMailContent(&mailTo, &senderConfigCount)

    if err := gomail.Send(__connect, m); err != nil {
        sd.SendCountLock.Lock()
        sd.SendFailCount++
        sd.SendCountLock.Unlock()
        writeLog(fmt.Sprintf("Error: Could not send email to %q: %v", mailTo, err))
    } else {
        sd.SendCountLock.Lock()
        sd.SendCount++
        sd.SendCountLock.Unlock()
        writeLog(
            fmt.Sprintf(
                "Count: %05d, Profile: %02d, Num: %03d, %s Send --> %q",
                sd.SendCount,
                senderConfigCount+1,
                mailToCount+1,
                sd.SendersProfile[senderConfigCount].MailFrom,
                mailTo))
    }
    m.Reset()
}

func (sd *SenderData) setMailContent () string {
    data, err := ioutil.ReadFile(sd.MailContentPath)
    if err != nil {
        log.Fatalf("***Error: %s", err)
    }
    log.Println("Open Mail Content File Successfully.")

    return string(data)
}

func (sd *SenderData) setMailList () []string {
    mailList, err := readLines(sd.MailListPath)
    if err != nil {
        log.Fatalf("***Error: %s", err)
    }
    log.Println("Open Mail List File Successfully.")

    return mailList
}

func (sd *SenderData) setSenderProfileConfig () []byte {
    sendersJson, err := ioutil.ReadFile(sd.ToolsArgs.SendProfileConfigPath)
    if err != nil {
        log.Fatalf("***Error: %s", err)
    }
    log.Println("Open Senders Config File Successfully.")

    return sendersJson
}

func (sd *SenderData) buildMailContent (mailTo *string, senderConfigCount *int) *gomail.Message {
    m := gomail.NewMessage(gomail.SetEncoding(gomail.Base64))
    m.SetHeader("From", m.FormatAddress(sd.SendersProfile[*senderConfigCount].MailFrom, sd.SendersProfile[*senderConfigCount].MailFromName))
    m.SetHeader("To", *mailTo)
    m.SetHeader("Subject", sd.Subject)
    m.SetHeader("X-Spam-Flag", "YES")
    m.SetHeader("X-Spam-Level", "*************")
    m.SetHeader("X-Spam-Status", "Yes, score=13.7 required=5.0 tests=BAYES_99,HS_INDEX_PARAM,HTML_MESSAGE,RCVD_IN_PBL,RCVD_IN_SORBS_DUL,RDNS_NONE,URIBL_AB_SURBL,URIBL_BLACK,URIBL_JP_SURBL,URIBL_SBL,URIBL_WS_SURBL autolearn=spam version=3.2.5")
    m.SetBody("text/html", sd.MailContent)

    return m
}

func writeLog(logContent string) {
    f, err := os.OpenFile(sendLogPath, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
    if err != nil {
        log.Fatalf("error opening file: %v", err)
    }
    defer f.Close()

    log.SetOutput(f)
    log.Println(logContent)
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
