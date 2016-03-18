package transcription

import (
	"io"
	"net/http"
	"net/smtp"
	"os"
	"os/exec"
	"strings"

	"github.com/hack4impact/audio-transcription-service/web"
)

// SendEmail connects to an email server at host:port, switches to TLS,
// authenticates on TLS connections using the username and password, and sends
// an email from address from, to address to, with subject line subject with message body.
func SendEmail(username string, password string, host string, port int, to []string, subject string, body string) error {
	from := username
	auth := smtp.PlainAuth("", username, password, host)

	// The msg parameter should be an RFC 822-style email with headers first,
	// a blank line, and then the message body. The lines of msg should be CRLF terminated.
	msg := []byte(msgHeaders(from, to, subject) + "\r\n" + body + "\r\n")
	addr := host + ":" + string(port)
	if err := smtp.SendMail(addr, auth, from, to, msg); err != nil {
		return err
	}
	return nil
}

func msgHeaders(from string, to []string, subject string) string {
	fromHeader := "From: " + from
	toHeader := "To: " + strings.Join(to, ", ")
	subjectHeader := "Subject: " + subject
	msgHeaders := []string{fromHeader, toHeader, subjectHeader}
	return strings.Join(msgHeaders, "\r\n")
}

// ConvertAudioIntoRequiredFormat converts encoded audio into the required format.
func ConvertAudioIntoRequiredFormat(fn string) error {
	// http://cmusphinx.sourceforge.net/wiki/faq
	// -ar 16000 sets frequency to required 16khz
	// -ac 1 sets the number of audio channels to 1
	cmd := exec.Command("ffmpeg", "-i", fn, "-ar", "16000", "-ac", "1", fn+".wav")
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

// DownloadFileFromURL locally downloads an audio file stored at url.
func DownloadFileFromURL(url string) error {
	// Taken from https://github.com/thbar/golang-playground/blob/master/download-files.go
	output, err := os.Create(FileNameFromURL(url))
	if err != nil {
		return err
	}
	defer output.Close()

	// Get file contents
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// Write the body to file
	_, err = io.Copy(output, response.Body)
	if err != nil {
		return err
	}

	return nil
}

// FileNameFromURL splits a URL by '/' to extract the file name.
func FileNameFromURL(url string) string {
	tokens := strings.Split(url, "/")
	fileName := tokens[len(tokens)-1]
	return fileName
}

// Task returns a task function for transcription using transcription functions.
func Task(jsonData web.transcriptionJobData) func() error {
	return func() error {
		fileName = transcription.FileNameFromURL(jsonData.AudioURL)
		if err = DownloadFileFromURL(fileName); err != nil {
			return err
		}
		if err = transcription.ConvertAudioIntoRequiredFormat(fileName); err != nil {
			return err
		}
	}
}
