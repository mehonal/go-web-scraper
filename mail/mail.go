package mail

import (
    "fmt"
    "strings"
    "datamin/config"
    "net/smtp"
)

func SendEmail( to []string, subject string, message string) ( err error ) { 
    auth := smtp.PlainAuth("", config.SMTP_EMAIL, config.SMTP_PASS, config.SMTP_SERVER)
    toHeader := strings.Join(to, ",")
    msg := []byte(
        "From: " + config.SMTP_EMAIL + "\r\n" +
        "To: " + toHeader + " \r\n" + 
        "Subject: " + subject + "\r\n" + 
        "\r\n" +
        message + "\r\n")
    fmt.Printf("%s\n", msg)
    err = smtp.SendMail(fmt.Sprintf("%s:%d", config.SMTP_SERVER, config.SMTP_PORT), auth, config.SMTP_EMAIL, to, msg)
    if err == nil { print("Email sent\n") }
    return err
}
