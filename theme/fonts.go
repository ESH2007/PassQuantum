package theme

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed fonts/IBMPlexSans-Regular.ttf
var ibmPlexSansRegular []byte

//go:embed fonts/IBMPlexSans-Medium.ttf
var ibmPlexSansMedium []byte

//go:embed fonts/IBMPlexSans-SemiBold.ttf
var ibmPlexSansSemiBold []byte

//go:embed fonts/IBMPlexMono-Regular.ttf
var ibmPlexMonoRegular []byte

//go:embed fonts/IBMPlexMono-Medium.ttf
var ibmPlexMonoMedium []byte

var (
	FontSansRegular  = fyne.NewStaticResource("IBMPlexSans-Regular.ttf", ibmPlexSansRegular)
	FontSansMedium   = fyne.NewStaticResource("IBMPlexSans-Medium.ttf", ibmPlexSansMedium)
	FontSansSemiBold = fyne.NewStaticResource("IBMPlexSans-SemiBold.ttf", ibmPlexSansSemiBold)
	FontMonoRegular  = fyne.NewStaticResource("IBMPlexMono-Regular.ttf", ibmPlexMonoRegular)
	FontMonoMedium   = fyne.NewStaticResource("IBMPlexMono-Medium.ttf", ibmPlexMonoMedium)
)
