package services

import (
	"crypto/md5"
	"encoding/hex"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GenerateUniqueStringFromObjectID(objectID primitive.ObjectID) string {
	hasher := md5.New()
	hasher.Write(objectID[:])
	hash := hex.EncodeToString(hasher.Sum(nil))
	return hash[:6]
}
