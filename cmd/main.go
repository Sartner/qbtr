package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	qbittorrent "github.com/autobrr/go-qbittorrent"
	"github.com/hekmon/transmissionrpc/v3"
	"github.com/spf13/cobra"
)

// Version information - will be injected during build
var (
	Version    = "dev"
	BuildDate  = "unknown"
	CommitHash = "unknown"

	// Application configuration
	qbURL         string
	qbUsername    string
	qbPassword    string
	qbAutoDelete  bool
	qbTorrentsDir string
	trURL         string
	trUsername    string
	trPassword    string
	trAutoStart   bool
	dryRun        bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "qbtr",
		Short: "Transfer completed torrents from qBittorrent to Transmission",
		Run: func(cmd *cobra.Command, args []string) {
			if dryRun {
				log.Println("Running in dry-run mode: won't delete torrents from qBittorrent, will add and delete from Transmission")
			}
			transferTorrents()
		},
	}

	// Version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("qbtr version %s\n", Version)
			fmt.Printf("Built on %s\n", BuildDate)
			fmt.Printf("Commit %s\n", CommitHash)
		},
	})

	// qBittorrent flags
	rootCmd.Flags().StringVar(&qbURL, "qb-url", "http://localhost:8080", "qBittorrent WebUI URL")
	rootCmd.Flags().StringVar(&qbUsername, "qb-username", "", "qBittorrent username")
	rootCmd.Flags().StringVar(&qbPassword, "qb-password", "", "qBittorrent password")
	rootCmd.Flags().BoolVar(&qbAutoDelete, "qb-auto-delete", false, "Automatically delete torrents from qBittorrent after transfer")

	// Transmission flags
	rootCmd.Flags().StringVar(&trURL, "tr-url", "http://localhost:9091/transmission/rpc", "Transmission RPC URL")
	rootCmd.Flags().StringVar(&trUsername, "tr-username", "", "Transmission username")
	rootCmd.Flags().StringVar(&trPassword, "tr-password", "", "Transmission password")
	rootCmd.Flags().BoolVar(&trAutoStart, "tr-auto-start", false, "Automatically start torrents in Transmission")

	// Directory flags
	rootCmd.Flags().StringVar(&qbTorrentsDir, "qb-torrents-dir", "", "Directory containing torrent files")

	// Dry run mode
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Test mode: won't delete torrents from qBittorrent, but will add and then delete from Transmission")

	// Mark required flags
	rootCmd.MarkFlagRequired("qb-torrents-dir")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func transferTorrents() {
	qbClient := connectToQBittorrent()
	trClient := connectToTransmission()
	completedTorrents := getCompletedTorrents(qbClient)

	for _, torrent := range completedTorrents {
		processTorrent(qbClient, trClient, torrent)
	}

	log.Println("Transfer completed")
}

func connectToQBittorrent() *qbittorrent.Client {
	qbClient := qbittorrent.NewClient(qbittorrent.Config{
		Host:     qbURL,
		Username: qbUsername,
		Password: qbPassword,
	})

	err := qbClient.Login()
	if err != nil {
		log.Fatalf("Failed to login to qBittorrent: %v", err)
	}
	log.Println("Successfully connected to qBittorrent")

	return qbClient
}

func getCompletedTorrents(qbClient *qbittorrent.Client) []qbittorrent.Torrent {
	torrents, err := qbClient.GetTorrents(qbittorrent.TorrentFilterOptions{
		Filter: qbittorrent.TorrentFilterCompleted,
	})
	if err != nil {
		log.Fatalf("Failed to get torrents from qBittorrent: %v", err)
	}

	log.Printf("Found %d completed torrents in qBittorrent", len(torrents))

	return torrents
}

func connectToTransmission() *transmissionrpc.Client {
	parsedURL, err := url.Parse(trURL)
	if err != nil {
		log.Fatalf("Failed to parse Transmission URL: %v", err)
	}

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	if trUsername != "" || trPassword != "" {
		parsedURL.User = url.UserPassword(trUsername, trPassword)
	}

	trConfig := transmissionrpc.Config{
		CustomClient: httpClient,
	}

	trClient, err := transmissionrpc.New(parsedURL, &trConfig)
	if err != nil {
		log.Fatalf("Failed to connect to Transmission: %v", err)
	}
	log.Println("Successfully connected to Transmission")

	return trClient
}

func processTorrent(qbClient *qbittorrent.Client, trClient *transmissionrpc.Client, torrent qbittorrent.Torrent) {
	torrentFilePath := filepath.Join(qbTorrentsDir, torrent.Hash+".torrent")
	if _, err := os.Stat(torrentFilePath); os.IsNotExist(err) {
		log.Printf("Torrent file not found for %s, skipping", torrent.Name)
		return
	}

	if tr, ok := addTorrentToTransmission(qbClient, trClient, torrentFilePath, torrent); ok {
		if dryRun {
			removeTorrentFromTransmission(trClient, tr)
		} else if qbAutoDelete {
			removeTorrentFromQBittorrent(qbClient, torrent)
		} else {
			log.Printf("Torrent added to Transmission but not removed from qBittorrent (qb-auto-delete is disabled): %s", torrent.Name)
		}
	}
}

func addTorrentToTransmission(qbClient *qbittorrent.Client, trClient *transmissionrpc.Client, torrentFilePath string, qbTorrent qbittorrent.Torrent) (transmissionrpc.Torrent, bool) {
	torrentB64, err := transmissionrpc.File2Base64(torrentFilePath)
	if err != nil {
		log.Printf("Failed to encode qbTorrent file %s: %v", torrentFilePath, err)
		return transmissionrpc.Torrent{}, false
	}

	paused := !trAutoStart
	if dryRun {
		paused = true
	}

	addOptions := transmissionrpc.TorrentAddPayload{
		MetaInfo:    &torrentB64,
		Paused:      &paused,
		DownloadDir: &qbTorrent.SavePath,
	}

	tr, err := trClient.TorrentAdd(context.Background(), addOptions)
	if err != nil {
		log.Printf("Failed to add qbTorrent %s to Transmission: %v", qbTorrent.Name, err)
		return transmissionrpc.Torrent{}, false
	}
	log.Printf("Added torrent to Transmission: %s (ID: %d)", *tr.Name, *tr.ID)
	if ok := addTrackersToTransmission(trClient, qbClient, &qbTorrent, &tr); !ok {
		return transmissionrpc.Torrent{}, false
	} else {
		log.Printf("Failed to add trackers to Transmission torrent: %s (ID: %d)", *tr.Name, *tr.ID)
	}
	return tr, true
}

// addTrackersToTransmission gets trackers from qbTorrent and adds them to the Transmission torrent
func addTrackersToTransmission(trClient *transmissionrpc.Client, qbClient *qbittorrent.Client, qbTorrent *qbittorrent.Torrent, trTorrent *transmissionrpc.Torrent) bool {
	trackers, ok := getQBTorrentTrackers(qbClient, qbTorrent)
	if !ok {
		return false
	}
	err := trClient.TorrentSet(context.Background(), transmissionrpc.TorrentSetPayload{
		IDs:         []int64{*trTorrent.ID},
		TrackerList: trackers,
	})
	if err != nil {
		log.Printf("Failed to update trackers for torrent %s in Transmission: %v", *trTorrent.Name, err)
		return false
	}
	return true
}

func getQBTorrentTrackers(qbClient *qbittorrent.Client, torrent *qbittorrent.Torrent) ([]string, bool) {
	trackers, err := qbClient.GetTorrentTrackers(torrent.Hash)
	if err != nil {
		log.Printf("Failed to get trackers for torrent %s: %v", torrent.Name, err)
		return nil, false
	}
	trackerUrls := make([]string, len(trackers))
	for _, tracker := range trackers {
		// if tracker.url not start with http, skip
		if tracker.Url == "" || tracker.Status == qbittorrent.TrackerStatusDisabled {
			continue
		}
		trackerUrls = append(trackerUrls, tracker.Url)
	}
	return trackerUrls, true
}

func removeTorrentFromTransmission(trClient *transmissionrpc.Client, tr transmissionrpc.Torrent) {
	torrentID := *tr.ID
	torrentName := *tr.Name

	err := trClient.TorrentRemove(context.Background(), transmissionrpc.TorrentRemovePayload{
		IDs:             []int64{torrentID},
		DeleteLocalData: false,
	})
	if err != nil {
		log.Printf("Failed to remove torrent %s from Transmission: %v", torrentName, err)
	} else {
		log.Printf("Removed torrent from Transmission in dry-run mode: %s", torrentName)
	}
}

func removeTorrentFromQBittorrent(qbClient *qbittorrent.Client, torrent qbittorrent.Torrent) {
	err := qbClient.DeleteTorrents([]string{torrent.Hash}, false)
	if err != nil {
		log.Printf("Failed to delete torrent %s from qBittorrent: %v", torrent.Name, err)
		return
	}
	log.Printf("Deleted torrent from qBittorrent: %s", torrent.Name)
}
