package main

import (
	"go-qbittorrent/qbittorrent"

	log "github.com/sirupsen/logrus"
)

func main() {
	qbt, err := qbittorrent.NewQbittorrentClient("http://127.0.0.1:8181", "admin", "adminadmin")
	if err != nil {
		log.Fatalf(err.Error())
	}

	err = qbt.AddTorrent(qbittorrent.AddTorrentInput{
		TorrentFilePath: qbittorrent.AddTorrentInputTorrentFilePath{Path: "/Users/ilyas/Projects/homelab/go-qbittorrent/go-qbt/test.torrent"},
		Tags:            "ec4e8436-9c8e-11ec-b909-0242ac120002",
	})
	if err != nil {
		log.Fatalf(err.Error())
	}

	torrents, err := qbt.GetTorrents(qbittorrent.GetTorrentsInput{
		Tag: "ec4e8436-9c8e-11ec-b909-0242ac120002",
	})
	if err != nil {
		log.Fatalf(err.Error())
	}
	for _, torrent := range torrents {
		log.Infof("%s %s", torrent.Name, torrent.Category)
	}

}
