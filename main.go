package main

import (
	lt "github.com/steeve/libtorrent-go"

	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path"
	"runtime"
	"time"
)

type TorrentState int

const (
	Queued_for_checking TorrentState = iota
	Checking_files
	Downloading_metadata
	Downloading
	Finished
	Seeding
	Allocating
	Checking_resume_data
)

// Libtorrent Alert notifications
const (
	error_notification        uint = 1 << iota
	peer_notification              = 1 << iota
	port_mapping_notification      = 1 << iota
	storage_notification           = 1 << iota
	tracker_notification           = 1 << iota
	debug_notification             = 1 << iota
	status_notification            = 1 << iota
	progress_notification          = 1 << iota
	ip_block_notification          = 1 << iota
	performance_warning            = 1 << iota
	dht_notification               = 1 << iota
	stats_notification             = 1 << iota
	rss_notification               = 1 << iota
	all_categories                 = ^uint32(0) >> 1
)

// Libtorrent Torrent_Handle Status flag
const (
	query_distributed_copies         uint = 1 << iota
	query_accurate_download_counters uint = 1 << iota
	query_last_seen_complete         uint = 1 << iota
	query_pieces                     uint = 1 << iota
	query_verified_pieces            uint = 1 << iota
	query_torrent_file               uint = 1 << iota
	query_name                       uint = 1 << iota
	query_save_path                  uint = 1 << iota
)

var SessionLookup = map[TorrentState]string{
	Queued_for_checking:  "Queued_for_checking",
	Checking_files:       "Checking_files",
	Downloading_metadata: "Downloading_metadata",
	Downloading:          "Downloading",
	Finished:             "Finished",
	Seeding:              "Seeding",
	Allocating:           "Allocating",
	Checking_resume_data: "Checking_resume_data",
}

type TorrentStatus struct {
	Name         string       `json:"name"`
	State        TorrentState `json:"state"`
	StateString  string       `json:"state_string"`
	Progress     float32      `json:"progress"`
	DownloadRate float32      `json:"download_rate"`
	UploadRate   float32      `json:"upload_rate"`
	NumPeers     int          `json:"num_peers"`
	NumSeeds     int          `json:"num_seeds"`
	TotalSeeds   int          `json:"total_seeds"`
	TotalPeers   int          `json:"total_peers"`
	HasMetadata  bool         `json:"has_metadata`
}

func NewTorrentStatus(torrentHandle lt.Torrent_handle) *TorrentStatus {
	tstatus := torrentHandle.Status()
	return &TorrentStatus{
		State:        TorrentState(tstatus.GetState()),
		StateString:  SessionLookup[TorrentState(tstatus.GetState())],
		Progress:     tstatus.GetProgress(),
		DownloadRate: float32(tstatus.GetDownload_rate()) / 1000,
		UploadRate:   float32(tstatus.GetUpload_rate()) / 1000,
		NumPeers:     tstatus.GetNum_peers(),
		TotalPeers:   tstatus.GetNum_incomplete(),
		NumSeeds:     tstatus.GetNum_seeds(),
		TotalSeeds:   tstatus.GetNum_complete(),
		HasMetadata:  tstatus.GetHas_metadata(),
	}
}

func SaveResumeData(alert lt.Alert) error {
	SaveResumeData := lt.SwigcptrSave_resume_data_alert(alert.Swigcptr())
	SaveResumeEntry := lt.SwigcptrEntry(SaveResumeData.GetResume_data().Swigcptr())
	log.Println(SaveResumeEntry.To_string())

	ioutil.WriteFile(path.Join(".", "torrents.resumedata"), []byte(lt.Bencode(SaveResumeEntry)), 0644)
	return nil
}

// taking inspiration from https://github.com/steeve/torrent2http/blob/master/torrent2http.go
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	ec := lt.NewError_code()
	// ec = lt.Lazy_bdecode(ioutil.ReadFile("torrents.resumedata"))
	// if ec.Value() != 0 {
	// 	log.Println(ec.Message())
	// }

	randomTorrent := lt.NewAdd_torrent_params()
	randomTorrent.SetUrl("magnet:?xt=urn:btih:F5483E44EBD64519D5FEACFC22F7373B03B4CB59&dn=the+good+lie+2014+720p+brrip+x264+yify&tr=udp%3A%2F%2F9.rarbg.com%3A2710%2Fannounce&tr=udp%3A%2F%2Fopen.demonii.com%3A1337")
	randomTorrent.SetSave_path(".")
	// torrentInfo := lt.NewTorrent_info("magnet:?xt=urn:btih:F5483E44EBD64519D5FEACFC22F7373B03B4CB59&dn=the+good+lie+2014+720p+brrip+x264+yify&tr=udp%3A%2F%2F9.rarbg.com%3A2710%2Fannounce&tr=udp%3A%2F%2Fopen.demonii.com%3A1337")
	// randomTorrent.SetTi(torrentInfo)

	torrentSession := lt.NewSession()
	torrentSession.Set_alert_mask(status_notification + storage_notification)
	torrentSession.Listen_on(lt.NewStd_pair_int_int(6900, 6999), ec)
	if ec.Value() != 0 {
		log.Println(ec.Message())
	}

	torrentHandle := torrentSession.Add_torrent(randomTorrent, ec)
	if ec.Value() != 0 {
		log.Println(ec.Message())
	}

	// shutdown idea
	saveChannel := make(chan bool, 0)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-saveChannel
		<-c
		log.Println("SHUTDOWN request received.")
		torrentSession.Pause()
		torrentHandle.Save_resume_data()
		<-saveChannel
		log.Println("received a reply, closing")
		os.Exit(1)
	}()

	go func() {
		for {
			torrentStatus := NewTorrentStatus(torrentHandle)
			log.Printf("\n%+v", torrentStatus)
			time.Sleep(5 * time.Second)
		}
	}()

	go func() {
		for {
			if torrentSession.Wait_for_alert(lt.Seconds(10)).Swigcptr() == 0 {
				log.Println("Alert timeout occurred!")
				continue
			}

			alert := torrentSession.Pop_alert()
			switch alert.What() {
			default:
				log.Printf("Alert: %#v", alert.What())
			case "metadata_received_alert":
				log.Println("Received Metadata!! finally!")
				torrentHandle.Save_resume_data()
			case "save_resume_data_alert":
				log.Println("Wrote Metadata!")
				_ = SaveResumeData(alert)
				saveChannel <- true
			case "save_resume_data_failed_alert":
				log.Println("Failed Metadata!")
				saveChannel <- false
			}
		}
	}()

	select {}
}
