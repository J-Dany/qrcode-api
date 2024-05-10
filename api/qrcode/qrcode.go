package qrcode

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"io"
	"qrcode/api/dto"
	"qrcode/database"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

////////////////////////////////////////////////
// Create
////////////////////////////////////////////////

func getSafeColor(val *color.RGBA) (color.RGBA, error) {
	if val == nil {
		return color.RGBA{}, fmt.Errorf("'val' is nil")
	}

	return color.RGBA{val.R, val.G, val.B, val.A}, nil
}

func Create(request *gin.Context) {
	user, _ := request.Get("user")
	var body dto.QrcodeCreateDto

	if err := request.ShouldBindWith(&body, binding.JSON); err != nil {
		request.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := validate.Struct(body); err != nil {
		request.JSON(400, gin.H{"error": err.Error()})
		return
	}

	log.Println(body)

	qrc, err := qrcode.New(body.Data)
	if err != nil {
		request.JSON(500, gin.H{"error": err.Error()})
		return
	}

	var options []standard.ImageOption

	if body.Options != nil {
		if body.Options.Transparent != nil && *body.Options.Transparent {
			options = append(options, standard.WithBgTransparent())
		}

		if body.Options.Size != nil {
			options = append(options, standard.WithQRWidth(*body.Options.Size))
		} else {
			options = append(options, standard.WithQRWidth(16))
		}

		if body.Options.Borders != nil {
			options = append(options, standard.WithBorderWidth(body.Options.Borders...))
		} else {
			options = append(options, standard.WithBorderWidth(0))
		}

		fg, err := getSafeColor(body.Options.ForegroundColor)
		if err == nil {
			options = append(options, standard.WithFgColor(fg))
		}

		bg, err := getSafeColor(body.Options.BackgroundColor)
		if err == nil {
			options = append(options, standard.WithBgColor(bg))
		}
	}

	filename := "./qrcode_" + body.Label + ".png"
	w, _ := standard.New(filename, options...)
	defer w.Close()
	defer os.Remove(filename)

	if err = qrc.Save(w); err != nil {
		request.JSON(500, gin.H{"error": err.Error()})
		return
	}

	bytes, _ := os.ReadFile(filename)

	// Save QR code to database
	database.InsertQrcode(user.(*database.User), &dto.InserQrcodeDto{
		Label: body.Label,
		Bytes: bytes,
		Data:  &body,
	})

	// Stream the PNG to the client
	request.Data(200, "image/png", bytes)
}

////////////////////////////////////////////////
// Get
////////////////////////////////////////////////

func GetQrcodeById(request *gin.Context) {
	requestedId := request.Param("id")

	id, _ := primitive.ObjectIDFromHex(requestedId)
	stream, err := database.GetQrcodeById(id)

	if err != nil {
		request.JSON(500, gin.H{"error": "Failed to get QR code"})
		return
	}

	request.Stream(func(w io.Writer) bool {
		_, err := io.Copy(w, stream)
		return err != nil
	})
}

func GetQrcodes(request *gin.Context) {
	user, _ := request.Get("user")
	qrcodes, err := database.GetQrcodesByUser(user.(*database.User))

	if err != nil {
		request.JSON(500, gin.H{"error": "Failed to get QR codes"})
		return
	}

	request.JSON(200, qrcodes)
}
