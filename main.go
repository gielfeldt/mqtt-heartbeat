package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	options := client.OptionsReader()
	fmt.Println("Connected to", options.Servers())
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
	panic(err)
}

func main() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		done <- true
	}()

	host := os.Getenv("MQTT_HOST")
	port := os.Getenv("MQTT_PORT")
	user := os.Getenv("MQTT_USER")
	pass := os.Getenv("MQTT_PASS")
	topic := os.Getenv("MQTT_TOPIC")
	interval, _ := time.ParseDuration(os.Getenv("HEARTBEAT_INTERVAL"))

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%s", host, port))
	opts.SetClientID("mqtt-heartbeat")
	opts.SetUsername(user)
	opts.SetPassword(pass)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	defer func() {
		token := client.Publish(topic+"/state", 1, true, "offline")
		token.Wait()
		fmt.Println("Offline")
		client.Disconnect(2000)
		fmt.Println("Disconnected")
	}()

	token := client.Publish(topic+"/state", 1, true, "online")
	token.Wait()
	fmt.Println("Online")

	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			token := client.Publish(topic+"/state", 1, true, "online")
			token.Wait()
			fmt.Println("Heartbeat")
		}
	}()

	<-done
	ticker.Stop()

}
