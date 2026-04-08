//go:build !nobiometric && !cgo

package main

import "fmt"

func main() {
	fmt.Println("biometric-demo requires CGO and OpenCV. Enable CGO and run with the native Windows biometric toolchain.")
	fmt.Println("Example (PowerShell):")
	fmt.Println("  $env:CGO_ENABLED=1")
	fmt.Println("  go run -tags customenv ./cmd/biometric-demo -camera 0 -threshold 0.97")
}
