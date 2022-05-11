package mongogo

import (
	"context"
	"fmt"
	"gin-server/internal/errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Connector struct {
	url      string
	client   *mongo.Client
	users    *mongo.Collection
	counters *mongo.Collection
}

type User struct {
	Id      int    `json:"id"`
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Friends []int  `json:"friends"`
}

type Counter struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

const (
	DATABASE string = "lesson31"
	USERS    string = "users"
	COUNTERS string = "counters"
)

func Init(url string) (Connector, error) {
	conn := Connector{}
	conn.url = "mongodb://" + url

	cliOptions := options.Client().ApplyURI(conn.url)
	client, err := mongo.Connect(context.TODO(), cliOptions)

	if err != nil {
		return Connector{}, &errors.InternarMongoError{Err: err}
	}
	if client.Ping(context.TODO(), nil) != nil {
		return Connector{}, &errors.InternarMongoError{Err: err}
	}

	conn.client = client
	conn.users = client.Database(DATABASE).Collection(USERS)
	conn.counters = client.Database(DATABASE).Collection(COUNTERS)

	return conn, nil
}

func (c *Connector) Disconnect() error {
	return c.client.Disconnect(context.TODO())
}

func (c *Connector) GetCounter(name string) (int, error) {
	filter := bson.D{{Key: "name", Value: name}}
	var result Counter

	err := c.counters.FindOne(context.TODO(), filter).Decode(&result)

	fmt.Println()

	if err == mongo.ErrNoDocuments {
		_, err := c.counters.InsertOne(
			context.TODO(),
			Counter{name, 1})

		if err != nil {
			return 0, &errors.InternarMongoError{Err: err}
		}
		return 1, nil
	} else if err != nil {
		return 0, &errors.InternarMongoError{Err: err}
	}

	return result.Value, nil
}

func (c *Connector) CounterPlusPlus(name string) error {
	filter := bson.D{{Key: "name", Value: name}}
	update := bson.D{{
		Key: "$inc", Value: bson.D{{Key: "value", Value: 1}},
	}}

	_, err := c.counters.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		return &errors.InternarMongoError{Err: err}
	}

	return nil
}

func (c *Connector) NewUser(name string, age int) (int, error) {
	userId, err := c.GetCounter("user_id")
	if err != nil {
		return 0, &errors.InternarMongoError{Err: err}
	}

	_, err = c.users.InsertOne(
		context.TODO(),
		User{userId, name, age, []int{}})

	if err != nil {
		return 0, &errors.InternarMongoError{Err: err}
	}

	err = c.CounterPlusPlus("user_id")
	if err != nil {
		return 0, err
	}

	return userId, nil
}

func (c *Connector) UpdateAge(user_id, newAge int) error {
	err := c.CheckIds([]int{user_id})
	if err != nil {
		return err
	}

	filter := bson.D{{Key: "id", Value: user_id}}
	update := bson.D{{
		Key:   "$set",
		Value: bson.D{{Key: "age", Value: newAge}},
	}}

	_, err = c.users.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		return &errors.InternarMongoError{Err: err}
	}

	return nil
}

func (c *Connector) AddFriend(user_id, friend_id int) error {
	err := c.CheckIds([]int{user_id, friend_id})
	if err != nil {
		return err
	}

	err = c.FriendExists(user_id, friend_id)
	if err != nil {
		return err
	}

	filter := bson.D{{Key: "id", Value: friend_id}}
	update := bson.D{{
		Key:   "$push",
		Value: bson.D{{Key: "friends", Value: user_id}},
	}}

	_, err = c.users.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		return &errors.InternarMongoError{Err: err}
	}

	return nil
}

func (c *Connector) FriendExists(user_id, friend_id int) error {
	filter := bson.D{
		{Key: "id", Value: friend_id},
		{Key: "friends", Value: user_id},
	}

	var result User
	err := c.users.FindOne(context.TODO(), filter).Decode(&result)

	if err == mongo.ErrNoDocuments {
		return nil
	} else if err != nil {
		return &errors.InternarMongoError{Err: err}
	} else {
		return &errors.FriendsExists{SourceId: user_id, TargetId: friend_id}
	}
}

func (c *Connector) DelFriend(user_id, friend_id int) error {
	filter := bson.D{{Key: "id", Value: user_id}}
	update := bson.D{{
		Key:   "$pull",
		Value: bson.D{{Key: "friends", Value: friend_id}},
	}}

	_, err := c.users.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		return &errors.InternarMongoError{Err: err}
	}

	return nil
}

func (c *Connector) DelUser(user_id int) (string, error) {
	user, err := c.GetUser(user_id)
	if err != nil {
		return "", err
	}

	filter := bson.D{{Key: "id", Value: user_id}}

	_, err = c.users.DeleteOne(context.TODO(), filter)
	if err != nil {
		return "", &errors.InternarMongoError{Err: err}
	}

	update := bson.D{{
		Key:   "$pull",
		Value: bson.D{{Key: "friends", Value: user_id}},
	}}

	_, err = c.users.UpdateMany(context.TODO(), bson.D{{}}, update)
	if err != nil {
		return "", &errors.InternarMongoError{Err: err}
	}

	return user.Name, nil
}

func (c *Connector) GetFriends(user_id int) ([]User, error) {
	user, err := c.GetUser(user_id)
	if err != nil {
		return []User{}, err
	}

	filter := bson.D{{
		Key: "id",
		Value: bson.D{{
			Key:   "$in",
			Value: user.Friends,
		}},
	}}

	cursor, err := c.users.Find(context.TODO(), filter)
	if err != nil {
		return []User{}, &errors.InternarMongoError{}
	}

	var result []User

	for cursor.Next(context.TODO()) {
		var friend User

		err := cursor.Decode(&friend)
		if err != nil {
			return []User{}, &errors.InternarMongoError{}
		}

		result = append(result, friend)
	}

	return result, nil
}

func (c *Connector) GetUser(user_id int) (User, error) {
	err := c.CheckIds([]int{user_id})
	if err != nil {
		return User{}, err
	}

	var result User
	filter := bson.D{{Key: "id", Value: user_id}}
	err = c.users.FindOne(context.TODO(), filter).Decode(&result)

	if err != nil {
		return User{}, &errors.InternarMongoError{Err: err}
	}

	return result, nil
}

func (c *Connector) CheckIds(user_ids []int) error {
	filter := bson.D{{
		Key: "id",
		Value: bson.D{{
			Key:   "$in",
			Value: user_ids,
		}},
	}}

	cursor, err := c.users.Find(context.TODO(), filter)
	if err != nil {
		return &errors.InternarMongoError{Err: err}
	}

	var result []bson.M
	cursor.All(context.TODO(), &result)

	if len(user_ids) == len(result) {
		return nil
	} else {
		return &errors.UndefinedIndexes{Indexes: user_ids}
	}
}
