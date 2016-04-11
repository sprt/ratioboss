package main

import (
	cryptorand "crypto/rand"
	"flag"
	"fmt"
	"log"
	"math"
	mathrand "math/rand"
	"time"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/tracker"
)

var (
	minSeeders, minLeechers    int
	baseDownSpeed, baseUpSpeed float64 // megabit/s
	downMargin, upMargin       float64 // percentage

	baseDownSpeedByte, baseUpSpeedByte int64 // byte/s
	curDownSpeedByte, curUpSpeedByte   int64 // byte/s

	peerID               [20]byte
	mi                   *metainfo.MetaInfo
	prevRespTime         time.Time
	started              bool
	downloaded, uploaded int64 // bytes
)

func init() {
	mathrand.Seed(time.Now().UTC().UnixNano())

	flag.IntVar(&minSeeders, "s", 15, "minimum number of seeders required to fake data")
	flag.IntVar(&minLeechers, "l", 15, "minimum number of leechers required to fake data")
	flag.Float64Var(&baseDownSpeed, "d", 0, "download speed (megabit/s)")
	flag.Float64Var(&downMargin, "dm", 0.25, "download speed margin (+/-%)")
	flag.Float64Var(&baseUpSpeed, "u", 0, "upload speed (megabit/s)")
	flag.Float64Var(&upMargin, "um", 0.25, "upload speed margin (+/-%)")

	log.SetFlags(0)
	log.SetOutput(new(logWriter))
}

func main() {
	flag.Parse()

	if baseDownSpeed == 0 || baseUpSpeed == 0 {
		log.Fatal("Must specify download speed and upload speed")
	}

	if flag.NArg() != 1 {
		log.Fatal("One torrent file required")
	}

	filename := flag.Arg(0)

	var err error
	mi, err = metainfo.LoadFromFile(filename)
	if err != nil {
		log.Fatalln("Error loading torrent:", err)
	}

	log.Println("Name:", mi.Info.Name)
	log.Println("Size:", mi.Info.TotalLength(), "bytes")

	baseDownSpeedByte = megabitToByte(baseDownSpeed)
	baseUpSpeedByte = megabitToByte(baseUpSpeed)

	cryptorand.Read(peerID[:])

	for {
		var event tracker.AnnounceEvent
		if !started {
			event = tracker.Started
		}

		if started {
			sincePrevResp := time.Since(prevRespTime).Seconds()
			downloaded += int64(float64(curDownSpeedByte) * sincePrevResp)
			downloaded = int64(math.Min(float64(downloaded), float64(mi.Info.TotalLength())))
			uploaded += int64(float64(curUpSpeedByte) * sincePrevResp)

			if downloaded == mi.Info.TotalLength() {
				event = tracker.Completed
			}
		}

		req := &tracker.AnnounceRequest{
			InfoHash:   *mi.Info.Hash,
			PeerId:     peerID,
			Downloaded: downloaded,
			Left:       uint64(mi.Info.TotalLength() - downloaded),
			Uploaded:   uploaded,
			Event:      event,
			NumWant:    -1,
		}

		hasResp := true
		var resp tracker.AnnounceResponse
		var respTime time.Time

		if !started {
			log.Print("Starting...")
			for {
				var err error
				resp, err = tracker.Announce(mi.Announce, req)
				if err == nil {
					respTime = time.Now()
					break
				}
				sleepFor := 10 * time.Second
				log.Print("Announce error, retrying in ", sleepFor, "...")
				time.Sleep(sleepFor)
			}
		} else {
			log.Printf("Announcing - downloaded: %d bytes, uploaded: %d bytes", downloaded, uploaded)
			if event != tracker.None {
				log.Println("Event:", event)
			}
			var err error
			resp, err = tracker.Announce(mi.Announce, req)
			if err != nil {
				hasResp = false
				log.Print("Announce error")
			} else {
				respTime = time.Now()
			}
		}

		if hasResp {
			log.Print("Seeders: ", resp.Seeders, ", leechers: ", resp.Leechers)
			if resp.Seeders < int32(minSeeders) || resp.Leechers < int32(minLeechers) {
				curDownSpeedByte = 0
				curUpSpeedByte = 0
				log.Print("Not enough peers, stalling")
			} else {
				curUpSpeedByte = baseUpSpeedByte + int64(randMargin(baseUpSpeedByte, upMargin))
				if mi.Info.TotalLength()-downloaded > 0 {
					curDownSpeedByte = baseDownSpeedByte + int64(randMargin(baseDownSpeedByte, downMargin))
				} else {
					curDownSpeedByte = 0
				}
			}
			log.Printf("Setting speeds - down: %.3f Mb/s, up: %.3f Mb/s\n",
				byteToMegabit(curDownSpeedByte),
				byteToMegabit(curUpSpeedByte))
		}

		started = true
		prevRespTime = respTime

		sleepFor := time.Duration(resp.Interval) * time.Second
		log.Println("Next announce at", time.Now().Add(sleepFor).Format(time.Kitchen))
		time.Sleep(sleepFor)
	}
}

func randMargin(n int64, margin float64) float64 {
	return (mathrand.Float64() - 0.5) * 2 * margin * float64(n)
}

func byteToMegabit(b int64) float64 {
	return float64(b) * 8 / 1e6
}

func megabitToByte(megabit float64) int64 {
	return int64(megabit / 8 * 1e6)
}

type logWriter struct{}

func (writer logWriter) Write(bytes []byte) (int, error) {
	return fmt.Print(time.Now().Format(time.Kitchen) + " " + string(bytes))
}
