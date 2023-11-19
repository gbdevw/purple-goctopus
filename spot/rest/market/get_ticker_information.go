package market

import "github.com/gbdevw/purple-goctopus/spot/rest/common"

// Asset Ticker Info
type AssetTickerInfo struct {
	// Ask array(<price>, <whole lot volume>, <lot volume>)
	Ask []string `json:"a"`
	// Bid array(<price>, <whole lot volume>, <lot volume>)
	Bid []string `json:"b"`
	// Last trade closed array(<price>, <lot volume>)
	Close []string `json:"c"`
	// Volume array(<today>, <last 24 hours>)
	Volume []string `json:"v"`
	// Volume weighted average price array(<today>, <last 24 hours>)
	VolumeAveragePrice []string `json:"p"`
	// Number of trades array(<today>, <last 24 hours>)
	Trades []int64 `json:"t"`
	// Low array(<today>, <last 24 hours>)
	Low []string `json:"l"`
	// High array(<today>, <last 24 hours>)
	High []string `json:"h"`
	// Today's opening price
	OpeningPrice string `json:"o"`
}

// GetTickerInformation request options
type GetTickerInformationRequestOptions struct {
	// Asset pairs to get data for.
	//
	// If nil or empty, all pairs are returned.
	Pairs []string `json:"pairs,omitempty"`
}

// GetTickerInformation response
type GetTickerInformationResponse struct {
	common.KrakenSpotRESTResponse
	// Ticker data by pair
	Result map[string]AssetTickerInfo `json:"result,omitempty"`
}

/*************************************************************************************************/
/* HELPER METHODS                                                                                */
/*************************************************************************************************/

// Get the price of the best ask out of this AssetTickerInfo
func (ati *AssetTickerInfo) GetAskPrice() string {
	return ati.Ask[0]
}

// Get the whole lot volume of the best ask out of this AssetTickerInfo
func (ati *AssetTickerInfo) GetAskWholeLotVolume() string {
	return ati.Ask[1]
}

// Get the lot volume of the best ask out of an AssetTickerInfo
func (ati *AssetTickerInfo) GetAskLotVolume() string {
	return ati.Ask[2]
}

// Get the price of the best bid out of this AssetTickerInfo
func (ati *AssetTickerInfo) GetBidPrice() string {
	return ati.Bid[0]
}

// Get the whole lot volume of the best bid out of this AssetTickerInfo
func (ati *AssetTickerInfo) GetBidWholeLotVolume() string {
	return ati.Bid[1]
}

// Get the lot volume of the best bid out of this AssetTickerInfo
func (ati *AssetTickerInfo) GetBidLotVolume() string {
	return ati.Bid[2]
}

// Get the price of the last trade out of this AssetTickerInfo
func (ati *AssetTickerInfo) GetLastTradePrice() string {
	return ati.Close[0]
}

// Get the lot volume of the last trade out of this AssetTickerInfo
func (ati *AssetTickerInfo) GetLastTradeLotVolume() string {
	return ati.Close[1]
}

// Get today's traded volume out of this AssetTickerInfo
func (ati *AssetTickerInfo) GetTodayVolume() string {
	return ati.Volume[0]
}

// Get past 24h traded volume out of this AssetTickerInfo
func (ati *AssetTickerInfo) GetPast24HVolume() string {
	return ati.Volume[1]
}

// Get today's volume average price out of this AssetTickerInfo
func (ati *AssetTickerInfo) GetTodayVolumeAveragePrice() string {
	return ati.VolumeAveragePrice[0]
}

// Get past 24h volume average price out of this AssetTickerInfo
func (ati *AssetTickerInfo) GetPast24HVolumeAveragePrice() string {
	return ati.VolumeAveragePrice[1]
}

// Get today's trade count out of this AssetTickerInfo
func (ati *AssetTickerInfo) GetTodayTradeCount() int64 {
	return ati.Trades[0]
}

// Get today's trade count out of this AssetTickerInfo
func (ati *AssetTickerInfo) GetPast24HTradeCount() int64 {
	return ati.Trades[1]
}

// Get today's low price out of this AssetTickerInfo
func (ati *AssetTickerInfo) GetTodayLow() string {
	return ati.Low[0]
}

// Get past 24h low price out of this AssetTickerInfo
func (ati *AssetTickerInfo) GetPast24HLow() string {
	return ati.Low[1]
}

// Get today's high price out of this AssetTickerInfo
func (ati *AssetTickerInfo) GetTodayHigh() string {
	return ati.High[0]
}

// Get past 24h high price out of this AssetTickerInfo
func (ati *AssetTickerInfo) GetPast24HHigh() string {
	return ati.High[1]
}

// Get today's opening price out of this AssetTickerInfo
func (ati *AssetTickerInfo) GetTodayOpen() string {
	return ati.OpeningPrice
}
