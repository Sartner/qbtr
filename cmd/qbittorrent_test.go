package main

import (
	"github.com/autobrr/go-qbittorrent"
	"log"
	"testing"
)

func TestQBGetTrackers(t *testing.T) {
	qbClient := qbittorrent.NewClient(qbittorrent.Config{
		Host:     "http://localhost:8080",
		Username: "admin",
		Password: "adminadmin",
	})

	err := qbClient.Login()
	if err != nil {
		log.Fatalf("Failed to login to qBittorrent: %v", err)
	}

	torrents := getCompletedTorrents(qbClient)
	for _, torrent := range torrents {
		log.Printf("Torrent Name: %s, Tracker: %s", torrent.Name, torrent.Tracker)
		log.Print(torrent)
		trackers, _ := qbClient.GetTorrentTrackers(torrent.Hash)
		for _, tracker := range trackers {
			log.Print(tracker.Url)
			log.Print(tracker.Status)
		}
	}

}
