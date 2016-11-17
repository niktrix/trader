package main

type Quote struct {
	Symbol               string  `json:"symbol"`
	DxSymbol             string  `json:"dxSymbol"`
	Exchange             string  `json:"exchange"`
	IsoExchange          string  `json:"isoExchange"`
	BzExchange           string  `json:"bzExchange"`
	Type                 string  `json:"type"`
	Name                 string  `json:"name"`
	Description          string  `json:"description"`
	Sector               string  `json:"sector"`
	Industry             string  `json:"industry"`
	Open                 float64 `json:"open"`
	High                 float64 `json:"high"`
	Low                  float64 `json:"low"`
	BidPrice             float64 `json:"bidPrice"`
	AskPrice             float64 `json:"askPrice"`
	AskSize              int     `json:"askSize"`
	BidSize              int     `json:"bidSize"`
	Size                 int     `json:"size"`
	BidTime              int64   `json:"bidTime"`
	AskTime              int64   `json:"askTime"`
	LastTradePrice       float64 `json:"lastTradePrice"`
	LastTradeTime        int64   `json:"lastTradeTime"`
	Volume               int     `json:"volume"`
	Change               float64 `json:"change"`
	ChangePercent        float64 `json:"changePercent"`
	PreviousClosePrice   float64 `json:"previousClosePrice"`
	FiftyDayAveragePrice float64 `json:"fiftyDayAveragePrice"`
	FiftyTwoWeekHigh     float64 `json:"fiftyTwoWeekHigh"`
	FiftyTwoWeekLow      float64 `json:"fiftyTwoWeekLow"`
	MarketCap            int64   `json:"marketCap"`
	SharesOutstanding    int64   `json:"sharesOutstanding"`
	Pe                   float64 `json:"pe"`
	ForwardPE            float64 `json:"forwardPE"`
	DividendYield        float64 `json:"dividendYield"`
	PayoutRatio          float64 `json:"payoutRatio"`
}
