package main

import (
	"encoding/binary"
	"bytes"
	"log"
)

// IntToHex converts an int64 to a byte array
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

//log err
func logErr(err error)  {
	if err != nil {
		log.Panic(err)
	}
}
