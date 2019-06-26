package main

import (
	"context"
	"fmt"
	"os"
	"time"
	"flag"

	proto "github.com/golang/protobuf/proto"
	logging "github.com/ipfs/go-log"
	gpx "github.com/tkrajina/gpxgo/gpx"
	application "github.com/upperwal/go-mesh/application"
	driver "github.com/upperwal/go-mesh/driver/eth"
	fpubsub "github.com/upperwal/go-mesh/pubsub"
	bootservice "github.com/upperwal/go-mesh/service/bootstrap"
	pubservice "github.com/upperwal/go-mesh/service/publisher"
)

func main() {
	logging.SetLogLevel("svc-bootstrap", "DEBUG")
	logging.SetLogLevel("application", "DEBUG")
	logging.SetLogLevel("svc-publisher", "DEBUG")
	logging.SetLogLevel("fpubsub", "DEBUG")
	logging.SetLogLevel("pubsub", "DEBUG")
	logging.SetLogLevel("eth-driver", "DEBUG")

	gpxFile := flag.String("gpx", "dummy.gpx", "GPX file path")
	ra := flag.String("ra", "", "Remote Access")
	bsNodes := flag.String("bs", "/ip4/127.0.0.1/udp/4000/quic/p2p/QmVbcMycaK8ni5CeiM7JRjBRAdmwky6dQ6KcoxLesZDPk9", "Bootstrap Nodes")

	app, err := application.NewApplication(context.Background(), nil, nil)
	if err != nil {
		fmt.Println(err)
	}

	ethDriver, err := driver.NewEthDriver(*ra)
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

	pubService.RegisterToPublish("GGN.BUS")

	gpxDataChan := make(chan gpx.GPXPoint, 10)
	go readGPX(gpxDataChan, *gpxFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	for {
		d := <-gpxDataChan

		lat := float32(d.GetLatitude())
		long := float32(d.GetLongitude())

		m := &VehiclePosition{
			Position: &Position{
				Latitude:  &lat,
				Longitude: &long,
			},
		}
		raw, err := proto.Marshal(m)
		if err != nil {
			fmt.Println(err)
		}
		pubService.PublishData("GGN.BUS", raw)
	}
	app.Wait()

}

func readGPX(out chan gpx.GPXPoint, gpxFile string) error {
	bus, err := gpx.ParseFile(gpxFile)
	if err != nil {
		return err
	}

	for _, t := range bus.Tracks {
		for _, ts := range t.Segments {
			for _, p := range ts.Points {
				out <- p
				fmt.Println("Point: Lat: ", p.GetLatitude(), "Long:", p.GetLongitude())

				time.Sleep(2 * time.Second)
			}
		}
	}
	return nil
}
