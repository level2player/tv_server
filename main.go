package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"strconv"

	"io/ioutil"
	"net/http"
	"time"
)

var (
	historyUrl = "http://www.zg.com/api/v1/kline"
	corsConfig = cors.Config{
		AllowOrigins:     []string{"*"},
		AllowOriginFunc:  func(origin string) bool { return true },
		AllowMethods:     []string{"GET"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}
)

func main() {
	r := gin.Default()
	routers := r.Group("/api/v1/udf")
	routers.Use(cors.New(corsConfig))
	routers.GET("/config", ConfigHandler)
	routers.GET("/symbols", SymbolsHandler)
	routers.GET("/history", HistoryHandler)
	err := r.Run(":8080")
	if err != nil {
		panic(err)
	}
}

func ConfigHandler(c *gin.Context) {
	result := struct {
		SupportsSearch       bool     `json:"supports_search"`
		SupportsGroupRequest bool     `json:"supports_group_request"`
		SupportsMarks        bool     `json:"supports_marks"`
		SupportedResolutions []string `json:"supported_resolutions"`
	}{SupportsSearch: true, SupportsGroupRequest: false, SupportsMarks: false, SupportedResolutions: []string{"30", "60", "240", "D"}}
	c.JSON(http.StatusOK, result)
}

func SymbolsHandler(c *gin.Context) {
	symbol := c.Query("symbol")
	result := struct {
		S                    string   `json:"s"`
		Name                 string   `json:"name"`
		Timezone             string   `json:"timezone"`
		Minmov               int      `json:"minmov"`
		Pricescale           int      `json:"pricescale"`
		Pointvalue           int      `json:"pointvalue"`
		HasIntraday          bool     `json:"has_intraday"`
		HasDaily             bool     `json:"has_daily"`
		HasWeeklyAndMonthly  bool     `json:"has-weekly-and-monthly"`
		HasNoVolume          bool     `json:"has-no-volume"`
		Ticker               string   `json:"ticker"`
		Description          string   `json:"description"`
		Type                 string   `json:"type"`
		DataStatus           string   `json:"data_status"`
		SupportedResolutions []string `json:"supported-resolutions"`
		IntradayMultipliers  []string `json:"intraday_multipliers"`
		SessionRegular       string   `json:"session-regular"`
		HasFractionalVolume  bool     `json:"has-fractional-volume"`
	}{S: "ok", Name: symbol,
		Timezone:             "Asia/Shanghai",
		Minmov:               1,
		Pricescale:           1000,
		Pointvalue:           1,
		IntradayMultipliers:  []string{"1", "5", "15", "30", "60", "D", "W"},
		HasIntraday:          true,
		HasDaily:             true,
		HasWeeklyAndMonthly:  true,
		Ticker:               symbol,
		Description:          symbol,
		Type:                 "bitcoin",
		DataStatus:           "streaming",
		SupportedResolutions: []string{"5", "10", "15", "30", "60", "120", "240", "D", "W"},
		SessionRegular:       "24x7"}
	c.JSON(http.StatusOK, result)
}

func HistoryHandler(c *gin.Context) {
	symbol := c.Query("symbol")
	from, _ := strconv.ParseInt(c.Query("from"), 10, 64)
	to, _ := strconv.ParseInt(c.Query("to"), 10, 64)
	resolution := c.Query("resolution")
	size, period := getTimeDiff(resolution, from, to)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get(fmt.Sprintf("%s?type=%s&size=%d&symbol=%s", historyUrl, period, size, symbol))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"S": "fail", "err_info": err.Error()})
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"S": "fail", "err_info": err.Error()})
		return
	}
	jsp := [][]interface{}{}
	err = json.Unmarshal(body, &jsp)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"S": "fail", "err_info": err.Error()})
		return
	}

	if len(jsp) == 0 {
		c.JSON(http.StatusOK, gin.H{"S": "no_data", "err_info": err.Error()})
	}
	tvResp := struct {
		S string    `json:"s"`
		T []int64   `json:"t"`
		C []float64 `json:"c"`
		O []float64 `json:"o"`
		H []float64 `json:"h"`
		L []float64 `json:"l"`
		V []float64 `json:"v"`
	}{
		S: "ok",
	}
	t := make([]int64, len(jsp))
	cl := make([]float64, len(jsp))
	o := make([]float64, len(jsp))
	h := make([]float64, len(jsp))
	l := make([]float64, len(jsp))
	v := make([]float64, len(jsp))

	for i, k := range jsp {
		t[i] = int64(k[0].(float64)) / 1000
		o[i], _ = strconv.ParseFloat(k[1].(string), 64)
		h[i], _ = strconv.ParseFloat(k[2].(string), 64)
		l[i], _ = strconv.ParseFloat(k[3].(string), 64)
		cl[i], _ = strconv.ParseFloat(k[4].(string), 64)
		v[i], _ = strconv.ParseFloat(k[5].(string), 64)
	}
	tvResp.T = t
	tvResp.C = cl
	tvResp.O = o
	tvResp.H = h
	tvResp.L = l
	tvResp.V = v

	c.JSON(http.StatusOK, tvResp)
}

func getTimeDiff(resolution string, from, to int64) (size int, period string) {
	diff := time.Unix(int64(to), 0).Sub(time.Unix(int64(from), 0))
	switch resolution {
	case "1":
		return int(diff.Minutes()), "1min"
	case "5":
		return int(diff.Hours()) * 12, "5min"
	case "15":
		return int(diff.Hours()) * 4, "15min"
	case "30":
		return int(diff.Hours()) * 2, "30min"
	case "60":
		return int(diff.Hours()), "hour"
	case "D":
		return int(diff.Hours() / 24), "day"
	case "W":
		return int(diff.Hours() / 24 / 7), "week"
	default:
		return 0, ""
	}
}
