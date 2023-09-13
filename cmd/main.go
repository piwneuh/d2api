// A simple example that uses the modules from the gsbot package and go-steam to log on
// to the Steam network.
//
// The command expects log on data, optionally with an auth code:
//
//     gsbot [username] [password]
//     gsbot [username] [password] [authcode]
package main

import (
	"fmt"
	"os"

	"github.com/paralin/go-steam"
	"github.com/paralin/go-steam/gsbot"
	"github.com/paralin/go-steam/protocol/steamlang"

	// "io/ioutil"
	// "math/rand"
	// "encoding/json"

	// "github.com/paralin/go-steam/netutil"
)

func main() {
	// d, err := ioutil.ReadFile("serverlist.json")
	// if err != nil {
	// 	panic(err)
	// }
	// println(string(d))

	// var addrs []*netutil.PortAddr
	// err = json.Unmarshal(d, &addrs)
	// if err != nil {
	// 	println(err.Error())
	// }
	// raddr := addrs[rand.Intn(len(addrs))]
	// println(raddr.String())

	if len(os.Args) < 3 {
		fmt.Println("gsbot example\nusage: \n\tgsbot [username] [password] [authcode]")
		return
	}

	// optional auth code
	authcode := ""
	if len(os.Args) > 3 {
		authcode = os.Args[3]
	}

	details := &gsbot.LogOnDetails{
		Username: os.Args[1], 
		Password: os.Args[2],
		AuthCode: authcode,
	}

	bot := gsbot.Default()
	client := bot.Client
	auth := gsbot.NewAuth(bot, details, "sentry.bin")
	debug, err := gsbot.NewDebug(bot, "debug")
	if err != nil {
		panic(err)
	}
	client.RegisterPacketHandler(debug)
	serverList := gsbot.NewServerList(bot, "serverlist.json")
	serverList.Connect()

	for event := range client.Events() {
		auth.HandleEvent(event)
		debug.HandleEvent(event)
		serverList.HandleEvent(event)

		switch e := event.(type) {
		case error:
			fmt.Printf("Error: %v", e)
		case *steam.LoggedOnEvent:
			client.Social.SetPersonaState(steamlang.EPersonaState_Online)
		}
	}
}