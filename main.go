package main

import (
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
	historyUrl = "https://api.hbdm.com/market/history/kline"
	corsConfig = cors.Config{
		AllowOrigins:     []string{"*"},
		AllowOriginFunc:  func(origin string) bool { return true },
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
)

func main() {
	r := gin.Default()
	routers := r.Group("/api/v1")
	routers.Use(cors.New(corsConfig))

	routers.GET("/udf/config", ConfigHandler)
	routers.GET("/udf/symbols", SymbolsHandler)
	routers.GET("/udf/history", HistoryHandler)
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
		Pricescale:           100,
		Pointvalue:           1,
		IntradayMultipliers:  []string{"1", "60", "D"},
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

	size, period := getDiff(resolution, from, to)

	u := fmt.Sprintf("%s?period=%s&size=%d&symbol=%s", historyUrl, period, size, symbol)
	resp, err := http.Get(u)
	if err != nil {

	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {

	}

	jsp := struct {
		Ch   string `json:"ch"`
		Data []struct {
			Amount float64 `json:"amount"`
			Close  float64 `json:"close"`
			Count  int     `json:"count"`
			High   int     `json:"high"`
			ID     int     `json:"id"`
			Low    int     `json:"low"`
			Open   float64 `json:"open"`
			Vol    int     `json:"vol"`
		} `json:"data"`
		Status string `json:"status"`
		Ts     int64  `json:"ts"`
	}{}

	err = json.Unmarshal(body, &jsp)
	if err != nil {

	}
	c.JSON(http.StatusOK, jsp)
}

func getDiff(resolution string, from, to int64) (size int, period string) {
	diff := time.Unix(int64(to), 0).Sub(time.Unix(int64(from), 0))
	if resolution == "D" {
		dHours := diff.Hours()
		if dHours <= 0 {
			return 1, "1day"
		} else {
			if (int(dHours) % 24) > 0 {
				return int(dHours)/24 + 1, "1day"
			} else {
				return int(dHours) / 24, "1day"
			}
		}
	} else {
		return 0, ""
	}

}
