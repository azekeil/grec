package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/azekeil/grec/internal/bot"
	"github.com/azekeil/grec/internal/commands"
	"github.com/azekeil/grec/internal/config"
	"github.com/azekeil/grec/internal/handler"
	"github.com/azekeil/grec/internal/self"

	"github.com/bwmarrin/discordgo"
)

func main() {
	v := config.ReadConfig()
	config := config.ParseConfig(v)

	// Setup help using the new embedded file approach
	files, err := commands.ParseAll()
	if err != nil {
		log.Fatalf("Failed to parse commands: %v", err)
	}
	help := self.BuildDocFromFiles(files)

	session, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

	// Register callback for the messageCreate events.
	session.AddHandler(handler.MakeMessageCreateHandlerFunc(help))

	// Register callback for the guildCreate events.
	session.AddHandler(handler.GuildCreate)

	// Open websocket after registering
	err = session.Open()
	if err != nil {
		panic(err)
	}

	// Wait here until CTRL-C or other term signal is received.
	log.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	log.Println("Caught signal! Stopping all recordings...")
	bot.StopAllRecordings(session)
	log.Println("Closing connection...")
	session.Close()
}
