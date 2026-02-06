package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// targetとportをstdInから受け付ける。
func recvTargetInfo() (string, string) {

	fmt.Println("waiting for your input(e.g. hostname port)...")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()

	target, port, b := strings.Cut(scanner.Text(), " ")
	if !b {
		// test my-tcp-server
		// target = "127.0.0.1"
		// port = "8080"

		// test google
		target = "www.google.com"
		port = "80"
	}

	fmt.Printf("%s:%s\n", target, port)

	fmt.Printf("address: target=%s, port=%s \n", target, port)

	return target, port
}

// ターゲットとポート番号のバリデーションのエントリ。
func validateInputArgs(target, port string) error {
	// hostname, ip-address, port
	if err := validateTargetInfo(target, port); err != nil {
		return err
	}

	return nil
}

// ターゲットとポート番号のバリデーション。
func validateTargetInfo(target, port string) error {
	// hostname
	if r, err := validate(matchHostNameString, target, inEligibleTarget); !r {
		// ip-address
		r, err = validate(matchIPAddressString, target, inEligibleTarget)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// port
	if _, err := validate(matchPortString, port, inEligiblePortNumber); err != nil {
		return err
	}

	return nil

}
func validate(m string, c string, e string) (bool, error) {
	r, err := regexp.MatchString(m, c)
	if !r {
		return r, errors.New(e)
	} else if err != nil {
		return r, err
	}

	return r, nil
}
