package main

import (
"fmt"
"time"

"gocv.io/x/gocv"
)

func test(api gocv.VideoCaptureAPI, idx int) {
cam, err := gocv.VideoCaptureDeviceWithAPI(idx, api)
if err != nil {
fmt.Printf("api=%d idx=%d open err: %v\n", api, idx, err)
return
}
defer cam.Close()

frame := gocv.NewMat()
defer frame.Close()

ok := false
for i := 0; i < 20; i++ {
if cam.Read(&frame) && !frame.Empty() {
ok = true
break
}
time.Sleep(80 * time.Millisecond)
}
fmt.Printf("api=%d idx=%d read_ok=%v rows=%d cols=%d\n", api, idx, ok, frame.Rows(), frame.Cols())
}

func main() {
apis := []gocv.VideoCaptureAPI{
gocv.VideoCaptureMSMF,
gocv.VideoCaptureDshow,
gocv.VideoCaptureAny,
gocv.VideoCaptureFFmpeg,
gocv.VideoCaptureGstreamer,
}
for _, api := range apis {
for idx := 0; idx <= 5; idx++ {
test(api, idx)
}
}
}
