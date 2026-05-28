package totp

import (
	"fmt"
	"image"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

func DecodeQRFromImage(img image.Image) (string, error) {
	src := gozxing.NewLuminanceSourceFromImage(img)
	reader := qrcode.NewQRCodeReader()

	hints := map[gozxing.DecodeHintType]interface{}{
		gozxing.DecodeHintType_TRY_HARDER: true,
	}
	if bmp, err := gozxing.NewBinaryBitmap(gozxing.NewHybridBinarizer(src)); err == nil {
		if result, err := reader.Decode(bmp, hints); err == nil {
			return result.GetText(), nil
		}
	}

	pureHints := map[gozxing.DecodeHintType]interface{}{
		gozxing.DecodeHintType_PURE_BARCODE: true,
	}
	if bmp, err := gozxing.NewBinaryBitmap(gozxing.NewHybridBinarizer(src)); err == nil {
		if result, err := reader.Decode(bmp, pureHints); err == nil {
			return result.GetText(), nil
		}
	}

	if bmp, err := gozxing.NewBinaryBitmap(gozxing.NewGlobalHistgramBinarizer(src)); err == nil {
		if result, err := reader.Decode(bmp, hints); err == nil {
			return result.GetText(), nil
		}
	}

	return "", fmt.Errorf("qr decode: could not decode QR code from image")
}

// DecodeQRToTOTP is a convenience that decodes a QR image and parses the
// resulting otpauth:// URI into TOTPParams in a single call.
func DecodeQRToTOTP(img image.Image) (*TOTPParams, error) {
	uri, err := DecodeQRFromImage(img)
	if err != nil {
		return nil, err
	}
	return ParseOTPAuthURI(uri)
}
