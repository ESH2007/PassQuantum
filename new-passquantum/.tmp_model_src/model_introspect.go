package main

import (
    "fmt"
    "image"
    "os"
    "gocv.io/x/gocv"
)

func main() {
    if len(os.Args) < 2 { panic("need model path") }
    path := os.Args[1]
    net := gocv.ReadNetFromONNX(path)
    if net.Empty() { panic("load failed") }
    defer net.Close()

    names := net.GetLayerNames()
    outIDs := net.GetUnconnectedOutLayers()
    fmt.Printf("layers=%d outIDs=%v\n", len(names), outIDs)

    outNames := make([]string,0,len(outIDs))
    for _, id := range outIDs {
        if id > 0 && id <= len(names) {
            outNames = append(outNames, names[id-1])
        }
    }
    fmt.Printf("outNames=%v\n", outNames)

    img := gocv.NewMatWithSize(192,192,gocv.MatTypeCV8UC3)
    defer img.Close()
    blob := gocv.BlobFromImage(img,1.0/127.5,image.Pt(192,192),gocv.NewScalar(127.5,127.5,127.5,0),true,false)
    defer blob.Close()
    net.SetInput(blob, "")

    for _, n := range outNames {
        m := net.Forward(n)
        fmt.Printf("forward[%s]: rows=%d cols=%d ch=%d total=%d type=%v\n", n, m.Rows(), m.Cols(), m.Channels(), m.Total(), m.Type())
        f, err := m.DataPtrFloat32()
        if err != nil {
            fmt.Printf("  DataPtrFloat32 err: %v\n", err)
        } else {
            fmt.Printf("  float count=%d\n", len(f))
        }
        m.Close()
    }
}
