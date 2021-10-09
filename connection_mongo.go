package connectionhelper

import (
	"context"
	"sync"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)
var clientInstance *mongo.Client

var clientInstanceError error

var mongoOnce sync.Once
const{
	CONNECTIONSTRING = "mongodb://localhost:8080"
	DB               = "db_issue_manager"
	ISSUES           = "col_issues"
)
func GetMongoClient() (*mongo.Client, error) {
	mongoOnce.Do(func() {
		clientOptions := options.Client().ApplyURI(CONNECTIONSTRING)
		// Connect to MongoDB
		client, err := mongo.Connect(context.TODO(), clientOptions)
		if err != nil {
			clientInstanceError = err
		}
		err = client.Ping(context.TODO(), nil)
		if err != nil {
			clientInstanceError = err
		}
		clientInstance = client
	})
	return clientInstance, clientInstanceError
}
