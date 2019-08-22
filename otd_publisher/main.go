package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	logging "github.com/ipfs/go-log"
	application "github.com/upperwal/go-mesh/application"
	driver "github.com/upperwal/go-mesh/driver/eth"
	fpubsub "github.com/upperwal/go-mesh/pubsub"
	bootservice "github.com/upperwal/go-mesh/service/bootstrap"
	pubservice "github.com/upperwal/go-mesh/service/publisher"
)

var (
	previousFileHash [16]byte
)

func main() {
	logging.SetLogLevel("svc-bootstrap", "DEBUG")
	logging.SetLogLevel("application", "DEBUG")
	//logging.SetLogLevel("svc-publisher", "DEBUG")
	logging.SetLogLevel("fpubsub", "DEBUG")
	logging.SetLogLevel("pubsub", "DEBUG")
	logging.SetLogLevel("eth-driver", "DEBUG")

	ra := flag.String("ra", "http://13.234.78.241:8501", "Remote Access")
	bsNodes := flag.String("bs", "/ip4/13.234.78.241/udp/4000/quic/p2p/QmVbcMycaK8ni5CeiM7JRjBRAdmwky6dQ6KcoxLesZDPk9", "Bootstrap Nodes")
	flag.Parse()

	app, err := application.NewApplication(context.Background(), nil, nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	ethDriver, err := driver.NewEthDriver(*ra, nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	bservice := bootservice.NewBootstrapService(false, "abc", []string{*bsNodes})
	pubService := pubservice.NewPublisherService(ethDriver)

	app.InjectService(bservice)
	app.InjectService(pubService)

	f := fpubsub.NewFilter(ethDriver)
	app.SetGossipPeerFilter(f)

	app.Start()

	time.Sleep(3 * time.Second)

	err = pubService.RegisterToPublish("GGN.BUS")
	if err != nil {
		panic(err)
	}

	rawOTDChan := make(chan []byte, 10)
	go readGPX(rawOTDChan)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	for {
		d, ok := <-rawOTDChan
		if !ok {
			return
		}

		pubService.PublishData("GGN.BUS", d)
	}

}

func readGPX(out chan []byte) {

	for {
		response, err := http.Get("https://otd.delhi.gov.in/api/realtime/VehiclePositions.pb?key=ZigTULDi3uDG5sPrHBKgDcJhzk2rZkjC")
		if err != nil {
			close(out)
			return
		}
		defer response.Body.Close()

		body, err := ioutil.ReadAll(response.Body)

		writeToFile(body)

		out <- body

		time.Sleep(10 * time.Second)
	}

	/* for _, t := range bus.Tracks {
		for _, ts := range t.Segments {
			for _, p := range ts.Points {
				out <- p
				fmt.Println("Point: Lat: ", p.GetLatitude(), "Long:", p.GetLongitude())

				time.Sleep(2 * time.Second)
			}
		}
	} */
}

func writeToFile(data []byte) error {
	nowHash := md5.Sum(data)
	if !bytes.Equal(nowHash[:], previousFileHash[:]) {
		previousFileHash = nowHash
		fileName := "data/otd_raw_" + strconv.FormatInt(time.Now().Unix(), 10)
		fmt.Println("Writting: " + fileName)
		return ioutil.WriteFile(fileName, data, 0644)
	}
	fmt.Println("Same File as previous: ", nowHash[:], previousFileHash)
	return nil
}
