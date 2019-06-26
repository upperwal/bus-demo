module github.com/upperwal/bus_demo

go 1.12

require (
	github.com/golang/protobuf v1.3.1
	github.com/ipfs/go-log v0.0.1
	github.com/tkrajina/gpxgo v1.0.1
	github.com/upperwal/go-mesh v1.0.0
	google.golang.org/grpc v1.20.1
)

replace github.com/upperwal/go-mesh v1.0.0 => ../go-mesh

replace github.com/libp2p/go-libp2p-pubsub v0.1.0 => ../go-libp2p-pubsub

replace github.com/upperwal/go-stun v0.0.1 => /Users/abhishek/go/src/github.com/upperwal/go-stun

replace github.com/upperwal/go-libp2p-quic-transport v0.3.0 => ../go-libp2p-quic-transport
