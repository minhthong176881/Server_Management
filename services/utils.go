package services

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/ssh"

	pbSM "github.com/minhthong176881/Server_Management/proto"
)

type Connection struct {
	*ssh.Client
	password string
}

func Connect(addr, user, password string) (*Connection, error) {
	if strings.Contains(addr, "127.0.0.1") || strings.Contains(addr, "localhost") {
		return nil, nil
	}
	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil }),
	}

	conn, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, err
	}
	return &Connection{conn, password}, nil
}

func SendEmail(message []string) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	from := os.Getenv("EMAIL")
	password := os.Getenv("EMAIL_PASSWORD")
	to := []string{
		"dominhthong99@gmail.com",
	}

	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")

	auth := smtp.PlainAuth("", from, password, smtpHost)

	t, _ := template.ParseFiles("template/email.html")
	var body bytes.Buffer
	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body.Write([]byte(fmt.Sprintf("Subject: Daily report \n%s\n\n", mimeHeaders)))

	t.Execute(&body, struct {
		Name    string
		Message []string
	}{
		Name:    "Thông đẹp trai siêu cấp vũ trụ",
		Message: message,
	})

	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, body.Bytes())
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Email sent successfully!")
}

func GetChangeLog(logs []*LogItem, changeLogs []*ChangeLogItem) []*ChangeLogItem {
	var startIndex, endIndex int
	var start, end string
	var recursive []*LogItem
	var countOff, countOn int
	if len(logs) <= 0 {
		return changeLogs
	}

	for i := 0; i < len(logs); i++ {
		if logs[i].Status == "Off" {
			countOff++
			start = logs[i].Time
			startIndex = i
			break
		}
	}
	if countOff == 0 {
		return changeLogs
	}

	for i := startIndex + 1; i < len(logs); i++ {
		if logs[i].Status == "On" {
			countOn++
			end = logs[i].Time
			endIndex = i
			break
		}
	}
	if countOn == 0 {
		end = logs[len(logs)-1].Time
		newChangeLog := &ChangeLogItem{}
		newChangeLog.Start = start
		newChangeLog.End = end
		newChangeLog.Total = CalculateTimeDiff(strings.Split(FormatTime(start), " ")[1], strings.Split(FormatTime(end), " ")[1])
		changeLogs = append(changeLogs, newChangeLog)
		return changeLogs
	}

	newChangeLog := &ChangeLogItem{}
	newChangeLog.Start = logs[startIndex].Time
	newChangeLog.End = logs[endIndex].Time
	newChangeLog.Total = CalculateTimeDiff(strings.Split(FormatTime(start), " ")[1], strings.Split(FormatTime(end), " ")[1])
	changeLogs = append(changeLogs, newChangeLog)
	for i := endIndex; i < len(logs); i++ {
		recursive = append(recursive, logs[i])
	}
	return GetChangeLog(recursive, changeLogs)
}

func FormatTime(date string) string {
	timestamp, _ := strconv.ParseInt(date, 10, 64)
	return time.Unix(timestamp, 0).String()
}

func CalculateTimeDiff(start string, end string) string {
	hourFormat := "15:04:05"
	t1, _ := time.Parse(hourFormat, start)
	t2, _ := time.Parse(hourFormat, end)
	return t2.Sub(t1).String()
}

func CheckValidTimeRange(start string, end string) bool {
	layout := "2006-01-02"
	t1, _ := time.Parse(layout, start)
	t2, _ := time.Parse(layout, end)
	return !strings.Contains(t2.Sub(t1).String(), "-")
}

func PbSMToService(server *pbSM.Server) (*Server, error) {
	oid, err := primitive.ObjectIDFromHex(server.Id)
	if err != nil {
		return nil, err
	}
	return &Server{
		ID: oid,
		Ip: server.Ip,
		Port: server.Port,
		Username: server.Username,
		Password: server.Password,
		Description: server.Description,
		Status: server.Status,
		Validate: server.Validate,
	}, nil
}

func ServiceToPbSM(server *Server) *pbSM.Server {
	return &pbSM.Server{
		Id: server.ID.Hex(),
		Ip: server.Ip,
		Port: server.Port,
		Username: server.Username,
		Password: server.Password,
		Description: server.Description,
		Status: server.Status,
		Validate: server.Validate,
	}
}
