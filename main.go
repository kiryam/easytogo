package main

import (
	"flag"
	"fmt"
	"io"
	"time"

	"github.com/jacobsa/go-serial/serial"
	"github.com/sigurn/crc8"
)

func main() {
	var portPath string
	var baudRate int
	flag.StringVar(&portPath, "port", "/dev/cu.SLAB_USBtoUART", "serial port")
	flag.IntVar(&baudRate, "baud", 38400, "baudrate")
	flag.Parse()

	options := serial.OpenOptions{
		PortName:        portPath,
		BaudRate:        uint(baudRate),
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 4,
	}

	port, err := serial.Open(options)
	if err != nil {
		panic(err)
	}

	defer port.Close()
	cli := NewEasyToGo(port)
	cli.SetupSender()

	time.Sleep(time.Second * 10)
}

type EasyToGo struct {
	readWriter io.ReadWriteCloser
}

func NewEasyToGo(readWriter io.ReadWriteCloser) *EasyToGo {
	e := EasyToGo{readWriter: readWriter}
	go e.startReader()
	return &e
}

func (e *EasyToGo) startReader() {
	p := make([]byte, 128)
	for {
		n, err := e.readWriter.Read(p)
		if err == io.EOF {
			break
		}
		fmt.Println(string(p[:n]))
	}
}

var preamble = []byte("PCCOM,06,")
var epilog = []byte{0x0D, 0x0A}

// write p, ???, baudrate=115200, inputMode=Arduplane, batteryUsed=FC, displayStatus=checked, videoStandard=PAL, positionX=120, ????, ID=15
const CONFIG_Ardupilot115200 = "p,00,03,03,01,01,00,78,00,0F"

func (e *EasyToGo) SetupSender() error {
	var packet []byte
	packet = append(packet, preamble...)                       // preamble
	packet = append(packet, []byte(CONFIG_Ardupilot115200)...) // settings

	table := crc8.MakeTable(crc8.CRC8_CDMA2000)
	crc := crc8.Checksum(packet, table)
	packet = append(packet, []byte(fmt.Sprintf("*%x", crc))...) // checksum

	packet = append(packet, epilog...) // epilog
	packet = append([]byte{0x24}, packet...)
	fmt.Printf("%s [% x]\r\n", packet, packet)
	n, err := e.readWriter.Write(packet)
	if err != nil {
		return err
	}
	fmt.Println("Wrote", n, "bytes.")
	return nil
}

func (e *EasyToGo) ReadSenderConfig() error {
	return nil
}
