package main

import (
    "fmt"
    "image"
    "os"
    "gocv.io/x/gocv"
)

func testModel(path string) {
    net := gocv.ReadNetFromONNX(path)
    if net.Empty() {
        fmt.Printf("%s => load failed\n", path)
        return
    }
    defer net.Close()
    net.SetPreferableBackend(gocv.NetBackendDefault)
    net.SetPreferableTarget(gocv.NetTargetCPU)

    img := gocv.NewMatWithSize(192, 192, gocv.MatTypeCV8UC3)
    defer img.Close()
    blob := gocv.BlobFromImage(img, 1.0/127.5, image.Pt(192, 192), gocv.NewScalar(127.5,127.5,127.5,0), true, false)
    defer blob.Close()
    net.SetInput(blob, "")
    out := net.Forward("")
    defer out.Close()
    floats, err := out.DataPtrFloat32()
    if err != nil {
        fmt.Printf("%s => forward err: %v rows=%d cols=%d channels=%d total=%d\n", path, err, out.Rows(), out.Cols(), out.Channels(), out.Total())
        return
    }
    fmt.Printf("%s => OK floats=%d rows=%d cols=%d channels=%d total=%d\n", path, len(floats), out.Rows(), out.Cols(), out.Channels(), out.Total())
}

func main() {
    if len(os.Args) < 2 {
        fmt.Println("need model path args")
        os.Exit(1)
    }
    for _, p := range os.Args[1:] {
        testModel(p)
    }
}
