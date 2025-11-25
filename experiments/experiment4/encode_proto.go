package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/events_proto"
	zstandard "github.com/klauspost/compress/zstd"
	"google.golang.org/protobuf/proto"
)

var EventsProto events_proto.TxEvents = events_proto.TxEvents{
	Events: []*events_proto.Event{
		{
			AtType: "/tm.GnoEvent",
			Type:   "ProfileFieldCreated",
			Attributes: []*events_proto.Attribute{
				{
					Key: "FieldType",
					Value: &events_proto.Attribute_StringValue{
						StringValue: "StringField",
					},
				},
			},
			PkgPath: proto.String("gno.land/r/demo/profile"),
		},
		{
			AtType: "at_type",
			Type:   "StorageDeposit",
			Attributes: []*events_proto.Attribute{
				{
					Key: "Deposit",
					Value: &events_proto.Attribute_Int64Value{
						Int64Value: 214600, // ugnot just for test this is a string
					},
				},
			},
			PkgPath: proto.String("gno.land/r/demo/profile"),
		},
		{
			AtType: "/tm.GnoEvent",
			Type:   "StorageDeposit",
			Attributes: []*events_proto.Attribute{
				{
					Key: "Deposit",
					Value: &events_proto.Attribute_DoubleValue{
						DoubleValue: 214600.05111, // ugnot just for test this is a string
					},
				},
			},
			PkgPath: proto.String("gno.land/r/demo/profile"),
		},
	},
}

func EncodeProto(txEvents *events_proto.TxEvents) ([]byte, error) {
	return proto.Marshal(txEvents)
}

func DecodeProto(data []byte) (*events_proto.TxEvents, error) {
	var txEvents events_proto.TxEvents
	err := proto.Unmarshal(data, &txEvents)
	if err != nil {
		return nil, err
	}
	return &txEvents, nil
}

func main() {
	encoded, err := EncodeProto(&EventsProto)
	if err != nil {
		log.Fatal(err)
	}

	hexEncoded := hex.EncodeToString(encoded)

	// Different ways to display the encoded data:
	fmt.Println("Raw bytes (decimal):", string(encoded))
	fmt.Println("Raw bytes (hex):", hexEncoded)
	fmt.Printf("Raw bytes (hex with spaces): %x\n", hexEncoded)
	fmt.Printf("Length: %d bytes\n", len(encoded))

	// try different compressions
	// Test with different compression levels
	for _, level := range []int{1, 2, 3, 10, 15, 18, 20} {
		encoderLevel := zstandard.EncoderLevelFromZstd(level)
		encoder, err := zstandard.NewWriter(nil, zstandard.WithEncoderLevel(encoderLevel))
		if err != nil {
			log.Fatal(err)
		}

		var buf bytes.Buffer
		encoder.Reset(&buf)
		_, err = encoder.Write(encoded)
		if err != nil {
			log.Fatal(err)
		}
		err = encoder.Close()
		if err != nil {
			log.Fatal(err)
		}

		compressed := buf.Bytes()
		fmt.Printf("Compressed (level %d): %d bytes (%.1f%% of original)\n", level, len(compressed), float64(len(compressed))/float64(len(encoded))*100)
	}
	// Final compression with level 22
	encoderLevel := zstandard.EncoderLevelFromZstd(22)
	encoder, err := zstandard.NewWriter(nil, zstandard.WithEncoderLevel(encoderLevel))
	if err != nil {
		log.Fatal(err)
	}

	var buf bytes.Buffer
	encoder.Reset(&buf)
	_, err = encoder.Write(encoded)
	if err != nil {
		log.Fatal(err)
	}
	err = encoder.Close()
	if err != nil {
		log.Fatal(err)
	}
	err = encoder.Close()
	if err != nil {
		log.Fatal(err)
	}
	compressed := buf.Bytes()
	fmt.Println("Final compressed bytes:", hex.EncodeToString(compressed))
	fmt.Printf("Final length: %d bytes\n", len(compressed))
	decoded, err := DecodeProto(encoded)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("decoded: ", decoded.String())
}
