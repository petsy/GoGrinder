package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"

	"github.com/finklabs/GoGrinder/gogrinder"
	"github.com/finklabs/GoGrinder/req"
	"github.com/xuyu/goredis"
)

const (
	RECORDS = 30
)

// initialize the GoGrinder
var gg = gogrinder.NewTest()

// define testcases using teststeps
func supercars_01_list(m *gogrinder.Meta, s gogrinder.Settings) {
	var mm *req.HttpMetric
	var resp map[string]interface{}
	c := req.NewDefaultClient()
	base := s["supercars_url"].(string)
	b := gg.NewBracket("01_01_supercars_list")
	r, err := http.NewRequest("GET", base+"/rest/supercars/", nil)
	if err != nil {
		m.Error += err.Error()
		mm = &req.HttpMetric{*m, 0, 0, 400}
	} else {
		resp, _, mm = req.DoJson(c, r, m)

		// assert record count
		count := len(resp["data"].([]interface{}))
		if count < RECORDS {
			mm.Error += "Error: less then 30 records in list response!"
		}
	}
	b.End(mm)
}

func supercars_02_read(m *gogrinder.Meta, s gogrinder.Settings) {
	var mm *req.HttpMetric
	var resp map[string]interface{}
	c := req.NewDefaultClient()
	b := gg.NewBracket("02_01_supercars_read")
	id := rand.Intn(RECORDS-1) + 1
	url := fmt.Sprintf("%s/rest/supercars/%05d", s["supercars_url"].(string), id)
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		m.Error += err.Error()
		mm = &req.HttpMetric{*m, 0, 0, 400}
	} else {
		resp, _, mm = req.DoJson(c, r, m)
		// assert record id
		i, err := strconv.Atoi(resp["_id"].(string))
		if err != nil || i != id {
			m.Error += "Error: retrived wrong record!"
		}
	}
	b.End(mm)
}

func supercars_03_create(m *gogrinder.Meta, s gogrinder.Settings) {
	var mm *req.HttpMetric
	var resp map[string]interface{}
	c := req.NewDefaultClient()
	base := s["supercars_url"].(string)
	newCar := map[string]interface{}{"name": "Ferrari Enzo", "country": "Italy",
		"top_speed": "218", "0-60": "3.4", "power": "650", "engine": "5998",
		"weight": "1365", "description": "The Enzo Ferrari is a 12 cylinder " +
		"mid-engine berlinetta named after the company's founder, Enzo Ferrari.",
		"image": "050.png"}
	b := gg.NewBracket("03_01_supercars_create")
	r, err := req.NewPostJsonRequest(base+"/rest/supercars/", newCar)
	if err != nil {
		m.Error += err.Error()
		mm = &req.HttpMetric{*m, 0, 0, 400}
	} else {
		resp, _, mm = req.DoJson(c, r, m)
		id := resp["_id"].(string)
		if i, err := strconv.Atoi(id); err != nil || i <= RECORDS {
			m.Error += "Error: something went wrong during new record creation!"
		} else {
			redis, err := goredis.Dial(
				&goredis.DialConfig{Address: s["redis_srv"].(string)})
			if err != nil {
				// is the redis server running? correct address?
				m.Error += err.Error()
			} else {
				redis.SAdd("supercars", id) // no way this can go wrong!
			}
		}
	}
	b.End(mm)
}

func supercars_04_update(m *gogrinder.Meta, s gogrinder.Settings) {
	var mm *req.HttpMetric
	//var resp map[string]interface{}
	c := req.NewDefaultClient()
	base := s["supercars_url"].(string)
	change := map[string]interface{}{"cylinders": "12", "name": "Ferrari Enzo",
		"country": "Italy", "top_speed": "218", "0-60": "3.4", "power": "650",
		"engine": "5998", "weight": "1365", "description": "The Enzo Ferrari " +
			"is a 12 cylinder mid-engine berlinetta named after the company's " +
			"founder, Enzo Ferrari.", "image": "050.png"}
	b := gg.NewBracket("04_01_supercars_update")
	redis, err := goredis.Dial(&goredis.DialConfig{Address: s["redis_srv"].(string)})
	if err != nil {
		// is the redis server running? correct address?
		m.Error += err.Error()
		mm = &req.HttpMetric{*m, 0, 0, 400}
	} else {
		id, err := redis.SPop("supercars")
		if err != nil {
			// probably run out of data - so it does not make sense to continue
			m.Error += err.Error()
			mm = &req.HttpMetric{*m, 0, 0, 400}
		} else {
			r, err := req.NewPutJsonRequest(base+"/rest/supercars/"+string(id),
				change)
			if err != nil {
				m.Error += err.Error()
				mm = &req.HttpMetric{*m, 0, 0, 400}
			} else {
				_, _, mm = req.DoJson(c, r, m)
				//tsUpdate(m, c, base + "/rest/supercars/" + string(id), change)

				// add the record back!
				redis.SAdd("supercars", string(id)) // no way this can go wrong!
			}
		}
	}
	b.End(mm)
}

func supercars_05_delete(m *gogrinder.Meta, s gogrinder.Settings) {
	var mm *req.HttpMetric
	c := req.NewDefaultClient()
	base := s["supercars_url"].(string)

	b := gg.NewBracket("05_01_supercars_delete")
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
	r, err := http.NewRequest("DELETE", base+"/rest/supercars/"+string(id), nil)
	if err != nil {
		m.Error += err.Error()
		mm = &req.HttpMetric{*m, 0, 0, 400}
	} else {
		_, _, mm = req.DoRaw(c, r, m)
	}
	b.End(mm)
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
