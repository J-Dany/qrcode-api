package database

import (
	"context"
	"time"
	"qrcode/api/dto"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
)

type Qrcode struct {
	Title  string             `json:"title"`
	ID     primitive.ObjectID `bson:"_id" json:"id"`
	UserID primitive.ObjectID `bson:"userId" json:"userId"`
	Qr     primitive.ObjectID `bson:"qr" json:"qr"`
}

func GetQrcodesByUser(user *User) ([]Qrcode, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := Client.Database(MongoDatabase).Collection("qrcodes")
	cursor, err := collection.Find(ctx, map[string]interface{}{"userId": user.ID})
	if err != nil {
		return nil, err
	}

	var qrcodes []Qrcode
	err = cursor.All(ctx, &qrcodes)
	if err != nil {
		return nil, err
	}

	if qrcodes != nil {
		return qrcodes, nil
	}

	return make([]Qrcode, 0), nil
}

func GetQrcodeById(id primitive.ObjectID) (*gridfs.DownloadStream, error) {
	bucket, err := gridfs.NewBucket(Client.Database(MongoDatabase))

	if err != nil {
		return nil, err
	}

	dStream, err := bucket.OpenDownloadStream(id)
	if err != nil {
		return nil, err
	}

	return dStream, nil
}

func InsertQrcode(user *User, dto *dto.InserQrcodeDto) (*Qrcode, error) {
	bucket, err := gridfs.NewBucket(Client.Database(MongoDatabase))

	if err != nil {
		return nil, err
	}

	upload, err := bucket.OpenUploadStream("qrcode")
	if err != nil {
		return nil, err
	}
	defer upload.Close()

	_, err = upload.Write(dto.Bytes)
	if err != nil {
		return nil, err
	}

	qrcode := &Qrcode{
		ID:     primitive.NewObjectID(),
		UserID: user.ID,
		Title:  dto.Label,
		Qr:     upload.FileID.(primitive.ObjectID), // Store the file ID from GridFS
	}

	// Insert the QR code document into the "qrcodes" collection
	collection := Client.Database(MongoDatabase).Collection("qrcodes")
	_, err = collection.InsertOne(context.Background(), qrcode)
	if err != nil {
		return nil, err
	}

	return qrcode, nil
}
