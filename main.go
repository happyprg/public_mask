package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	"github.com/tylerb/graceful"
)

func main() {

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	e.GET(Path, NewIndexHandler().Handle)

	e.Server.Addr = ":8080"
	e.StdLogger.Println("server start with port -" + e.Server.Addr)
	if err := graceful.ListenAndServe(e.Server, 5*time.Second); err != nil {
		panic(err)
	}

}

const Path = "/"

type Handler struct {
}

type Stores struct {
	Address string              `json:"address"`
	Count   int64               `json:"count"`
	Stores  SortableStoreResult `json:"stores"`
}
type StoreResult struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Addr string `json:"addr"`
	//판매처 유형[약국: '01', 우체국: '02', 농협: '03']
	Stype string  `json:"type"`
	Lat   float32 `json:"lat"`
	Lng   float32 `json:"lng"`
	//string($YYYY/MM/DD HH:mm:ss)
	StockAt comparableTimeStr `json:"stock_at"`
	//재고 상태[100개 이상(녹색): 'plenty' / 30개 이상 100개미만(노랑색): 'some' / 2개 이상 30개 미만(빨강색): 'few' / 1개 이하(회색): 'empty']
	RemainStat stockString       `json:"remain_stat"`
	CreatedAt  comparableTimeStr `json:"created_at"`
}

type comparableTimeStr string

type stockString string

func (s stockString) toStockStatus() int {
	switch s {
	case "plenty":
		return 3
	case "some":
		return 2
	case "few":
		return 1
	default:
		return 0
	}
}

func (s comparableTimeStr) compare(other comparableTimeStr) int {

	t1, err := time.Parse("2006/01/02 15:04:05", string(s))
	if err != nil {
		return 0
	}

	t2, err := time.Parse("2006/01/02 15:04:05", string(other))

	if t1.After(t2) == true {
		return 1
	}
	return -1

}

type SortableStoreResult []StoreResult

func (p SortableStoreResult) Len() int {
	return len(p)
}

func (p SortableStoreResult) Less(i, j int) bool {

	firstOrderResult := p[i].RemainStat.toStockStatus() > p[j].RemainStat.toStockStatus()
	if firstOrderResult != true {
		secondOrderResult := p[i].StockAt.compare(p[j].StockAt)
		return secondOrderResult == 1
	}
	return firstOrderResult
}

func (p SortableStoreResult) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

type Params struct {
	Address string
	HasOnly bool
}

func (h *Handler) Handle(c echo.Context) error {
	addr := c.QueryParam("addr")
	if len(addr) < 1 {
		return c.String(http.StatusBadRequest, "missing addr query param")
	}
	filter := c.QueryParam("filter")
	if filter == "" {
		filter = "true"
	}
	hasStock, err := strconv.ParseBool(filter)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	searchURL := fmt.Sprintf("%s?address=%s", "https://8oi9s0nnth.apigw.ntruss.com/corona19-masks/v1/storesByAddr/json", url.QueryEscape(addr))
	response, err := http.Get(searchURL)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("searchURL - %s, err - %s", searchURL, err.Error()))
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err.Error())
		return c.String(http.StatusInternalServerError, fmt.Sprintf(" err - %s", err.Error()))
	}
	stores := Stores{}
	err = json.Unmarshal(body, &stores)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf(" err - %s", err.Error()))
	}

	var filteredResult SortableStoreResult
	for _, item := range stores.Stores {
		if hasStock == true && item.RemainStat.toStockStatus() == 0 {
			continue
		}
		filteredResult = append(filteredResult, item)
	}

	sort.Sort(filteredResult)
	resData, err := json.Marshal(filteredResult)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf(" err - %s", err.Error()))
	}
	return c.String(http.StatusOK, string(resData))
}

func NewIndexHandler() *Handler {
	return &Handler{}
}
