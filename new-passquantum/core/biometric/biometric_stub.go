//go:build nobiometric

package biometric

import (
	"fmt"
)

type FaceDetector struct{}
type FaceMesh struct{}
type Pipeline struct {
	Threshold float32
}

func NewFaceDetector(modelPath string) (*FaceDetector, error) {
	return nil, fmt.Errorf("biometric support disabled in this build")
}

func (fd *FaceDetector) Close() {}

func NewFaceMesh(modelPath string) (*FaceMesh, error) {
	return nil, fmt.Errorf("biometric support disabled in this build")
}

func (fm *FaceMesh) Close() {}

func NewPipeline(blazeFaceModelPath, faceMeshModelPath string, threshold float32) (*Pipeline, error) {
	return nil, fmt.Errorf("biometric support disabled in this build")
}

func (p *Pipeline) Close() {}

func (p *Pipeline) Run(frame any) ([]Landmark, bool, error) {
	return nil, false, fmt.Errorf("biometric support disabled in this build")
}

func (p *Pipeline) RunFrame(frame any) ([]Landmark, bool, error) {
	return nil, false, fmt.Errorf("biometric support disabled in this build")
}

func (p *Pipeline) BackendName() string {
	return "disabled"
}
