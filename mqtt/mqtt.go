package mqtt

import (
	"flag"
	"fmt"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

var client MQTT.Client

func init() {
	broker := flag.String("broker", "tcp://115.28.142.203:1883", "")
	password := flag.String("password", "", "The password (optional)")
	user := flag.String("user", "", "The User (optional)")
	id := flag.String("id", "testgoid", "The ClientID (optional)")
	cleansess := flag.Bool("clean", false, "Set Clean Session (default false)")
	// payload := flag.String("message", "hhhhhhhhhhh", "The message text to publish (default empty)")
	store := flag.String("store", ":memory:", "The Store Directory (default use memory store)")
	flag.Parse()

	opts := MQTT.NewClientOptions()
	opts.AddBroker(*broker)
	opts.SetClientID(*id)
	opts.SetUsername(*user)
	opts.SetPassword(*password)
	opts.SetCleanSession(*cleansess)
	if *store != ":memory:" {
		opts.SetStore(MQTT.NewFileStore(*store))
	}

	client = MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
}

func RemindEatMedicine(payload string) {
	fmt.Println(payload)
	topic := flag.String("topic", "/raspberry/medicine", "The topic name to/from which to publish/subscribe")
	qos := flag.Int("qos", 0, "The Quality of Service 0,1,2 (default 0)")
	token := client.Publish(*topic, byte(*qos), false, payload)
	if token.Error() != nil {
		fmt.Println(token.Error())
	}
}
