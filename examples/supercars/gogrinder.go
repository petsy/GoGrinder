package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"

	"github.com/finklabs/GoGrinder/gogrinder"
	"github.com/finklabs/GoGrinder/http"
	"github.com/xuyu/goredis"
)

const (
	RECORDS = 30
)

// initialize the GoGrinder
var gg = gogrinder.NewTest()

// instrument teststeps
var tsList = gg.Teststep("01_01_supercars_list", http.GetJson)
var tsRead = gg.Teststep("02_01_supercars_read", http.GetJson)
var tsCreate = gg.Teststep("03_01_supercars_create", http.PostJson)
var tsUpdate = gg.Teststep("04_01_supercars_update", http.PutJson)
var tsDelete = gg.Teststep("05_01_supercars_delete", http.DeleteRaw)

// define testcases using teststeps
func supercars_01_list(m gogrinder.Meta, s gogrinder.Settings) {
	c := http.NewDefaultClient()
	base := s["supercars_url"].(string)
	resp := tsList(m, c, base+"/rest/supercars/").(http.ResponseJson).Json

	// assert record count
	count := len(resp["data"].([]interface{}))
	if count < RECORDS {
		m.Error += "Error: less then 30 records in list response!"
	}
}

func supercars_02_read(m gogrinder.Meta, s gogrinder.Settings) {
	c := http.NewDefaultClient()
	base := s["supercars_url"].(string)
	id := rand.Intn(RECORDS-1) + 1
	url := fmt.Sprintf("%s/rest/supercars/%05d", base, id)
	resp := tsRead(m, c ,url).(http.ResponseJson).Json

	// assert record id
	i, err := strconv.Atoi(resp["_id"].(string))
	if err != nil || i != id {
		m.Error += "Error: retrived wrong record!"
	}
}

func supercars_03_create(m gogrinder.Meta, s gogrinder.Settings) {
	c := http.NewDefaultClient()
	base := s["supercars_url"].(string)
	newCar := map[string]interface{}{"name": "Ferrari Enzo", "country": "Italy",
		"top_speed": "218", "0-60": "3.4", "power": "650", "engine": "5998", "weight": "1365",
		"description": "The Enzo Ferrari is a 12 cylinder mid-engine berlinetta named " +
			"after the company's founder, Enzo Ferrari.", "image": "050.png"}

	resp := tsCreate(m, c, base+"/rest/supercars/", newCar).(http.ResponseJson).Json
	id := resp["_id"].(string)
	if i, err := strconv.Atoi(id); err != nil || i <= RECORDS {
		m.Error += "Error: something went wrong during new record creation!"
		return
	}

	redis, err := goredis.Dial(&goredis.DialConfig{Address: s["redis_srv"].(string)})
	if err != nil {
		// is the redis server running? correct address?
		m.Error += err.Error()
		return
	}

	redis.SAdd("supercars", id) // no way this can go wrong!
}

func supercars_04_update(m gogrinder.Meta, s gogrinder.Settings) {
	c := http.NewDefaultClient()
	base := s["supercars_url"].(string)
	change := map[string]interface{}{"cylinders": "12", "name": "Ferrari Enzo",
		"country": "Italy", "top_speed": "218", "0-60": "3.4", "power": "650", "engine": "5998",
		"weight": "1365", "description": "The Enzo Ferrari is a 12 cylinder mid-engine " +
			"berlinetta named after the company's founder, Enzo Ferrari.", "image": "050.png"}

	redis, err := goredis.Dial(&goredis.DialConfig{Address: s["redis_srv"].(string)})
	if err != nil {
		// is the redis server running? correct address?
		m.Error += err.Error()
		return
	}

	id, err := redis.SPop("supercars")
	if err != nil {
		// probably run out of data - so it does not make sense to continue
		m.Error += err.Error()
		return
	}
	tsUpdate(m, c, base+"/rest/supercars/"+string(id), change)

	// add the record back!
	redis.SAdd("supercars", string(id)) // no way this can go wrong!
}

func supercars_05_delete(m gogrinder.Meta, s gogrinder.Settings) {
	c := http.NewDefaultClient()
	base := s["supercars_url"].(string)

	redis, err := goredis.Dial(&goredis.DialConfig{Address: s["redis_srv"].(string)})
	if err != nil {
		// is the redis server running? correct address?
		m.Error += err.Error()
		return
	}

	id, err := redis.SPop("supercars")
	if err != nil {
		// probably run out of data - so it does not make sense to continue
		m.Error += err.Error()
		return
	}
	tsDelete(m, c, base+"/rest/supercars/"+string(id))
}

// this is my endurance test scenario
func endurance() {
	// use the tests with the loadmodel config (json file)
	gg.Schedule("supercars_01_list", supercars_01_list)
	gg.Schedule("supercars_02_read", supercars_02_read)
	gg.Schedule("supercars_03_create", supercars_03_create)
	gg.Schedule("supercars_04_update", supercars_04_update)
	gg.Schedule("supercars_05_delete", supercars_05_delete)
}

// this is my baseline test scenario
func baseline() {
	// use the tests with a explicit configuration
	gg.DoIterations(supercars_01_list, 5, 0, false)
	gg.DoIterations(supercars_02_read, 5, 0, false)
	gg.DoIterations(supercars_03_create, 5, 0, false)
	gg.DoIterations(supercars_04_update, 5, 0, false)
	gg.DoIterations(supercars_05_delete, 5, 0, false)
}

func init() {
	// register the scenarios defined above
	gg.Testscenario("endurance", endurance)
	gg.Testscenario("baseline", baseline)
	// register the testcases as scenarios to allow single execution mode
	gg.Testscenario("supercars_01_list", supercars_01_list)
	gg.Testscenario("supercars_02_read", supercars_02_read)
	gg.Testscenario("supercars_03_create", supercars_03_create)
	gg.Testscenario("supercars_04_update", supercars_04_update)
	gg.Testscenario("supercars_05_delete", supercars_05_delete)
}

func main() {
	//gg.AddReportPlugin(http.NewHttpMetricReporter())
	err := gogrinder.GoGrinder(gg)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
