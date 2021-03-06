package bot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"os"
	"time"
)

type DBWorker struct {
	DBSession *mgo.Session
	DBName    string
}

type dbLog struct {
	Date   time.Time
	Text   string
	Module string
	Guild  string
}

type GuildData struct {
	ID          string
	WeatherCity string
	NewsCounty  string
	Language    string
	Timezone    int
	EmbedColor  int
}

type GuildsMap map[string]*GuildData

func NewDBSession(dbname string) *DBWorker {
	session, err := mgo.Dial(os.Getenv("MONGO_CONN"))
	if err != nil {
		fmt.Printf("Mongo connection error: %v", err)
	}
	count, err := session.DB("dtbot").C("logs").Count()
	if err != nil {
		fmt.Println("DB_ERR: ", err)
	}
	fmt.Printf("Mongo connected\nLogs in base: %v\n", count)
	return &DBWorker{DBSession: session, DBName: dbname}
}

// InitGuilds initialize guilds in database
func (db *DBWorker) InitGuilds(sess *discordgo.Session, conf *Config) GuildsMap {
	var data = make(GuildsMap)
	var loaded, initialized = 0, 0
	for _, guild := range sess.State.Guilds {
		count, err := db.DBSession.DB(db.DBName).C("guilds").Find(bson.M{"id": guild.ID}).Count()
		if err != nil {
			fmt.Println("Mongo: ", err)
		}
		if count == 0 {
			newData := &GuildData{
				ID:          guild.ID,
				WeatherCity: conf.Weather.City,
				NewsCounty:  conf.News.Country,
				Language:    conf.General.Language,
				Timezone:    conf.General.Timezone,
				EmbedColor:  conf.General.EmbedColor,
			}
			db.DBSession.DB(db.DBName).C("guilds").Insert(newData)
			data[guild.ID] = newData
			initialized++
		} else {
			var newData = &GuildData{}
			db.DBSession.DB(db.DBName).C("guilds").Find(bson.M{"id": guild.ID}).One(newData)
			if err != nil {
				fmt.Println("Mongo: ", err)
				continue
			}
			data[guild.ID] = newData
			loaded++
		}
	}
	fmt.Printf("Guilds loaded [%v], initialized [%v]\n", loaded, initialized)
	return data
}

// Log saves log in database
func (db *DBWorker) Log(module, guildID, text string, ) {
	db.DBSession.DB(db.DBName).C("logs").Insert(dbLog{Date: time.Now(), Text: text, Module: module, Guild: guildID})
}

// LogGet returns last N log rows
func (db *DBWorker) LogGet(count int) []dbLog {
	var log = make([]dbLog, count)
	db.DBSession.DB(db.DBName).C("logs").Find(nil).Sort("-$natural").Limit(count).All(&log)
	return log
}

func (db *DBWorker) Guilds() *mgo.Collection {
	return db.DBSession.DB(db.DBName).C("guilds")
}
