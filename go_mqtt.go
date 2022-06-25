package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	_ "github.com/mattn/go-sqlite3"
	"github.com/robfig/cron/v3"
	"github.com/slack-go/slack"
	"github.com/spf13/viper"
	"mqtt-server/avatars"
	"mqtt-server/services"
	"os"
	"os/signal"
	"strconv"
	"strings"
)

type JsonMessage struct {
	Id           string `json:"id"`
	Confidence   string `json:"confidence"`
	Name         string `json:"name"`
	Manufacturer string `json:"manufacturer"`
	Type         string `json:"type"`
	Retained     string `json:"retained"`
	Timestamp    string `json:"timestamp"`
	Version      string `json:"version"`
}

type UserAddRequest struct {
	Name       string `json:"name"`
	AvatarName string `json:"avatar_name"`
	MacAddress string `json:"mac_address"`
}

var mapJson = map[string]bool{}
var api *slack.Client

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	options := client.OptionsReader()
	fmt.Println("Connected to: ", options.Servers())
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connection lost: %v", err)
}

func createKnownUser(client mqtt.Client, mac, alias string) {
	text := fmt.Sprintf("%s %s", mac, alias)
	token := client.Publish("monitor/setup/ADD STATIC DEVICE", 2, false, text)
	token.Wait()
}

func subscribeToMonitor(client mqtt.Client) {
	topic := viper.GetString("monitor.topic")
	token := client.Subscribe(topic, 1, nil)
	token.Wait()
	fmt.Printf("Subscribed to topic: %s\n", topic)
}

func subscribeToUsers(client mqtt.Client) {
	topic := viper.GetString("users_topic")
	token := client.Subscribe(topic, 1, nil)
	token.Wait()
	fmt.Printf("Subscribed to topic: %s\n", topic)
}

func sendSlackMessageWithImage(filePath, content string) error {
	fmt.Println("uploading: ", filePath)
	f, err := os.Open(filePath)
	params := slack.FileUploadParameters{
		Title:    "Presence right now",
		Reader:   bufio.NewReader(f),
		Filename: "presence",
		Content:  content,
		Channels: []string{viper.GetString("slack.channel_id")},
	}
	// can use the file to delete and re-upload for maintaining a single image
	_, err = api.UploadFile(params)
	if err != nil {
		return fmt.Errorf("could not upload file with params %v %w", params, err)
	}
	return nil
}

func notifyPresence(db *services.DB) error {
	presence, err := db.PresenceService.GetPresence()
	var presenceImages []avatars.PresenceImage

	for _, p := range presence {
		presenceImages = append(presenceImages, avatars.PresenceImage{
			ImageName: p.AvatarName,
			Present:   p.IsPresent,
		})
	}
	path, err := avatars.CreateAvatar(presenceImages)
	if err != nil || len(path) == 0 {
		return fmt.Errorf("error creating avatar %s\n", err)
	}
	err = sendSlackMessageWithImage(path, "Current Office Presence")
	if err != nil {
		return fmt.Errorf("error sending slack %s\n", err)
	}
	return nil
}

var messageHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
	db := services.GetDBInstance()

	switch topic := msg.Topic(); topic {
	case viper.GetString("users_topic"):
		{
			fmt.Println(viper.GetString("users_topic"))
			var j UserAddRequest
			err := json.Unmarshal(msg.Payload(), &j)
			if err != nil {
				fmt.Printf("could not Unmarshal %s \n\n", err)
				return
			}
			err = db.UserService.CreateUser(j.Name, j.AvatarName, j.MacAddress)
			if err != nil {
				fmt.Printf("could not CreateUser %s \n", err)
				return
			}
			createKnownUser(client, j.MacAddress, j.Name)
		}
	default:
		{
			isMonitorTopic := strings.HasPrefix(topic, strings.TrimSuffix(viper.GetString("monitor.topic"), "#")) && !strings.HasSuffix(topic, "rssi") && !strings.HasSuffix(topic, "status")

			if isMonitorTopic {

				fmt.Println("is monitored!", viper.GetString("monitor.topic"))

				var jsonMessage JsonMessage
				err := json.Unmarshal(msg.Payload(), &jsonMessage)
				if err != nil {
					fmt.Printf("could not Unmarshal %s \n", err)
					return
				}

				user, err := db.UserService.GetUserByMac(jsonMessage.Id)
				if err != nil {
					fmt.Printf("could not get user by mac %s \n", err)
					return
				}
				//string to int
				confidence, err := strconv.Atoi(jsonMessage.Confidence)
				if confidence == 100 {
					fmt.Printf("%s is here", jsonMessage.Id)
					err := db.PresenceService.UpsertPresence(user.UID, true)
					if err != nil {
						fmt.Printf("could not upsert presence %s \n", err)
						return
					}
					err = notifyPresence(db)
					if err != nil {
						// send Slack message to channel general
						_, _, _ = api.PostMessage(viper.GetString("slack.channel_id"), slack.MsgOptionText(fmt.Sprintf("%s just entered the office :wave:", user.Name), false))
					}

				} else if confidence == 0 {
					fmt.Printf("%s is gone", jsonMessage.Id)
					err := db.PresenceService.UpsertPresence(user.UID, false)
					if err != nil {
						fmt.Printf("could not upsert presence %s \n", err)
						return
					}
					err = notifyPresence(db)
					if err != nil {
						// send Slack message to channel general
						_, _, _ = api.PostMessage(viper.GetString("slack.channel_id"), slack.MsgOptionText(fmt.Sprintf("%s just left the office :door:", user.Name), false))
					}

				}
				res := ""
				// iterate over map
				for key, value := range mapJson {
					res += fmt.Sprintf("%s: %t\n", key, value)
				}

			} else {
				fmt.Printf("not subscribed to topic %s.\n", topic)

			}
		}

	}

}

func main() {

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	fmt.Println("Config file successfully read")

	api = slack.New(viper.GetString("slack.token"))

	var broker = viper.GetString("broker.address")
	var port = viper.GetString("broker.port")
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%s", broker, port))
	opts.SetUsername(viper.GetString("broker.username"))
	opts.SetPassword(viper.GetString("broker.password"))

	db := services.GetDBInstance()

	opts.SetDefaultPublishHandler(messageHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	subscribeToMonitor(client)
	subscribeToUsers(client)

	_, err = api.AuthTest()
	if err != nil {
		panic(err)
	}

	cr := cron.New()
	_, err = cr.AddFunc("CRON_TZ=Asia/Jerusalem @midnight", func() {
		err := db.PresenceService.ResetPresence()
		if err != nil {
			fmt.Printf("could not reset presence %s\n", err)
		}
	})

	if err != nil {
		panic(err)
	}
	cr.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c
	fmt.Println("Disconnecting")
	client.Disconnect(250)
	db.CloseConnection()
	cr.Stop()

}
