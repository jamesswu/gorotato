package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gocolly/colly"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var db *mongo.Database

// data from job__tbody
type Action struct {
	Name     string `selector:".skill p strong" json:"name" bson:"name"`
	Image    string `selector:".skill .job_skill_icon img" json:"img" bson:"img"`
	Acquired string `selector:".jobclass p" json:"acquired" bson:"acquired"`
	Type     string `selector:".classification" json:"type" bson:"type"`
	Cast     string `selector:".cast" json:"cast" bson:"cast"`
	Recast   string `selector:".recast" json:"recast" bson:"recast"`
	MPCost   string `selector:".cost" json:"mpcost" bson:"mpcost"`
	Range    string `selector:".distant_range" json:"range" bson:"range"`
	Radius   string `selector:".distant_range" json:"radius" bson:"radius"`
	Effect   string `selector:".content" json:"effect" bson:"effect"`
	Revision string `selector:".content .update_text p" json:"revision" bson:"revision"`
}

func LoadEnv() error {
	if err := godotenv.Load(); err != nil {
		fmt.Println(("no .env file found"))
	}
	return nil
}

func InitDB() error {
	uri := os.Getenv(("MONGODB_URI"))
	if uri == "" {
		fmt.Println(("you must set your 'MONGODB_URI' environment variable"))
	}
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}

	db = client.Database("xivrotation")
	return nil
}

func CloseDB() error {
	return db.Client().Disconnect(context.Background())
}

func GetDBColleciton(c string) *mongo.Collection {
	return db.Collection(c)
}

func main() {

	err := LoadEnv()
	if err != nil {
		panic(err)
	}

	err = InitDB()
	if err != nil {
		panic(err)
	}

	defer CloseDB()

	c := colly.NewCollector()

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	coll := GetDBColleciton("Actions")

	c.OnHTML("tr[id^='pve_action__']", func(e *colly.HTMLElement) {
		action := Action{}
		e.Unmarshal(action)
		action.Name = e.ChildText(".skill p strong")
		action.Image = e.DOM.Find("img").AttrOr("src", "")
		action.Acquired = e.ChildText(".jobclass p")
		action.Type = e.ChildText(".classification")
		action.Cast = e.ChildText(".cast")
		action.Recast = e.ChildText(".recast")
		action.MPCost = e.ChildText(".cost")
		action.Range = e.ChildText(".distant_range")
		action.Radius = e.ChildText(".distant_range")
		action.Effect = e.DOM.Find(".content").Children().Remove().End().Text()
		action.Revision = e.ChildText(".content .update_text p")
		_, err := coll.InsertOne(context.TODO(), action)
		if err != nil {
			fmt.Println("insert error")
		} else {
			fmt.Println("insert successful")
		}

	})

	c.Visit("https://na.finalfantasyxiv.com/jobguide/gunbreaker/")
}
