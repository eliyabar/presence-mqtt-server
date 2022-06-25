package avatars

import (
	"bufio"
	"fmt"
	"github.com/slack-go/slack"
	"os"
	"testing"
)

var api *slack.Client

func TestAvatarCreate(t *testing.T) {
	a := []PresenceImage{
		{Present: true, ImageName: "e.png"},
	}
	avatar, err := CreateAvatar(a)
	if err != nil {
		t.Fatalf("could not Unmarshal %s \n", err)
	}

	api = slack.New("")
	f, err := os.Open("")
	_, _ = api.AuthTest()

	fmt.Println(avatar)
	params := slack.FileUploadParameters{
		Title:    "Current",
		Reader:   bufio.NewReader(f),
		Filename: "presence",
		Channels: []string{""},
	}
	file, err := api.UploadFile(params)
	if err != nil {
		t.Fatalf("could not Unmarshal %s \n", err)
		return
	}
	fmt.Printf("%v \n\n", file)
}

func TestSlackFIle(t *testing.T) {
	api = slack.New("")
	_, err := api.AuthTest()
	if err != nil {
		panic(err)
	}
	f, err := os.Open("./test.png")

	params := slack.FileUploadParameters{
		Title:    "Batman Example",
		Reader:   bufio.NewReader(f),
		Filename: "test.png",
		Channels: []string{""},
	}
	file, err := api.UploadFile(params)
	if err != nil {
		t.Fatalf("could not Unmarshal %s \n", err)
		return
	}
	fmt.Printf("%v \n\n", file)
}
