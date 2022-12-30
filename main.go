package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/tarm/serial"
)

var programmer *serial.Port

func main() {
	var err error
	portFlag := flag.String("p", "", "Port name")
	baudFlag := flag.Int("b", 115200, "Baud rate to use")
	readFlag := flag.Bool("r", false, "Read the EEPROM")
	unlockFlag := flag.Bool("u", false, "Unlock the EEPROM before clearing or write")
	clearFlag := flag.Bool("c", false, "Clear the EEPROM before programming")
	writeFlag := flag.String("w", "", "Write the EEPROM with the given data")
	lockFlag := flag.Bool("l", false, "Lock the EEPROM after clearing or write")
	flag.Parse()
	if *portFlag == "" {
		flag.Usage()
		return
	}
	baud := *baudFlag
	if baud <= 0 {
		baud = 9600
		return
	}

	programmer, err = serial.OpenPort(&serial.Config{
		Name:        *portFlag,
		Baud:        baud,
		ReadTimeout: time.Second * 2,
		Size:        8,
		Parity:      serial.ParityNone,
		StopBits:    serial.Stop1,
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer programmer.Close()

	// clear Port buffer
	buf := make([]byte, 255)
	for {
		l, _ := programmer.Read(buf)
		if l == 0 {
			break
		}
	}

	if err = checkConnection(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if *readFlag {
		if err = readEEPROM(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	if *unlockFlag {
		if err = unlockEEPROM(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	if *clearFlag {
		if err = clearEEPROM(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	if *writeFlag != "" {
		if err = writeEEPROM(*writeFlag); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	if *lockFlag {
		if err = lockEEPROM(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	fmt.Println("\nDone.")
}

func writeEEPROM(filename string) error {
	fmt.Println("Writing EEPROM...")
	dat, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	packet := make([]byte, 6)
	packet[0] = '<'
	packet[1] = 'W'
	packet[5] = '='

	pos := 0
	for addr, data := range dat {
		packet[2] = byte(addr >> 8)
		packet[3] = byte(addr & 0xFF)
		packet[4] = data

		if pos == 0 {
			fmt.Printf("\n%02x%02x: ", packet[2], packet[3])
		}
		if pos == 8 {
			fmt.Print(" ")
		}
		fmt.Printf(" %02x", packet[4])
		pos++
		if pos == 16 {
			pos = 0
		}

		if _, err = programmer.Write(packet); err != nil {
			return err
		}
		err = readResponse()
		if err != nil {
			return err
		}
	}

	return nil
}

func readEEPROM() error {
	fmt.Println("Reading EEPROM...")
	if _, err := programmer.Write([]byte("<R=")); err != nil {
		return err
	}
	return readResponse()
}

func lockEEPROM() error {
	fmt.Println("Locking EEPROM...")
	if _, err := programmer.Write([]byte("<L=")); err != nil {
		return err
	}
	return readResponse()
}

func unlockEEPROM() error {
	fmt.Println("Unlocking EEPROM...")
	if _, err := programmer.Write([]byte("<U=")); err != nil {
		return err
	}
	return readResponse()
}

func clearEEPROM() error {
	fmt.Println("Clearing EEPROM...")
	if _, err := programmer.Write([]byte("<C=")); err != nil {
		return err
	}
	return readResponse()
}

func readData() ([]byte, error) {
	count := 0
	var err error
	var data []byte
	dataStarted := false
	for {
		buf := make([]byte, 1)
		count, err = programmer.Read(buf)
		if err != nil {
			return nil, err
		}
		// fmt.Println("Read: ", count, " bytes: ", buf)
		if count > 0 {
			for i := 0; i < count; i++ {
				if buf[i] == '<' {
					dataStarted = true
				} else if buf[i] == '=' {
					return data, nil
				} else if dataStarted {
					data = append(data, buf[i])
				}
			}
		}
	}
}

func checkConnection() error {

	// send HELLO
	if _, err := programmer.Write([]byte("<H=")); err != nil {
		return err
	}

	// read response
	data, err := readData()
	if err != nil {
		return err
	}

	if string(data) != "HELLO" {
		return errors.New("Invalid response")
	}

	return nil
}

func readResponse() error {
	// read response
	pos := 0
	for {
		bData, err := readData()
		if err != nil {
			return errors.New("\nError reading data: " + err.Error())
		}
		data := string(bData)
		if data == "OK" {
			break
		} else if data == "ERROR" {
			return errors.New("\nOperation failed")
		}

		// Incoming data format: Address:Data
		arr := strings.Split(data, ":")
		if len(arr) != 2 {
			return errors.New("\nInvalid response")
		}
		if pos == 0 {
			fmt.Print("\n", arr[0], ": ")
		}
		if pos == 8 {
			fmt.Print(" ")
		}
		fmt.Print(" ", arr[1])
		pos++
		if pos == 16 {
			pos = 0
		}
	}
	return nil
}
