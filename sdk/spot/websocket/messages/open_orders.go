package messages

import (
	"encoding/json"
	"fmt"
)

// Data of a openOrders message from the websocket server
type OpenOrders struct {
	// Open orders as an array of maps where keys are the order ids and alues the orders
	Orders []map[string]OrderInfo
	// Sequence ID used to ensure no message is lost
	Sequence SequenceId
	// Channel name
	ChannelName string
}

// Custom JSON marhsaller for OpenOrders wich produces the same payloads as the API.
func (oo OpenOrders) MarshalJSON() ([]byte, error) {
	return json.Marshal([]interface{}{
		oo.Orders,
		oo.ChannelName,
		oo.Sequence,
	})
}

// Custom JSON unmarshaller for OpenOrders
func (oo *OpenOrders) UnmarshalJSON(data []byte) error {
	// Prepare an array of objects to parse the payload
	tmp := []interface{}{
		&[]map[string]OrderInfo{}, // Orders
		"",                        // Channel name
		&SequenceId{},             // Sequence Id object
	}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return fmt.Errorf("failed to parse data as OrderInfo: %w", err)
	}
	// Encode target
	oo.ChannelName = tmp[1].(string)
	oo.Orders = *tmp[0].(*[]map[string]OrderInfo)
	oo.Sequence = *tmp[2].(*SequenceId)
	return nil
}

// Description for a Order Info
type OrderInfoDescription struct {
	// Asset pair
	Pair string `json:"pair,omitempty"`
	// Optional - position ID (if applicable)
	PositionId string `json:"position,omitempty"`
	// Order direction (buy/sell). Cf. SideEnum.
	Type string `json:"type,omitempty"`
	// Order type. Cf. OrderTypeEnum
	OrderType string `json:"ordertype,omitempty"`
	// Limit or trigger price depending on order type
	Price string `json:"price,omitempty"`
	// Limit price for stop/take orders
	Price2 string `json:"price2,omitempty"`
	// Amount of leverage
	Leverage string `json:"leverage,omitempty"`
	// Textual order description
	OrderDescription string `json:"order,omitempty"`
	// Conditional close order description
	CloseOrderDescription string `json:"close,omitempty"`
}

// Description of a close order.
type CloseOrderInfo struct {
	// Close order type. Cf. OrderTypeEnum
	OrderType string `json:"ordertype,omitempty"`
	// Limit or trigger price depending on order type
	Price string `json:"price,omitempty"`
	// Limit price for stop/take orders
	Price2 string `json:"price2,omitempty"`
	// Comma delimited list of order flags
	//
	// viqc = volume in quote currency (not currently available), fcib = prefer fee in base currency,
	// fciq = prefer fee in quote currency, nompp = no market price protection, post = post only order
	// (available when ordertype = limit).
	OrderFlags string `json:"oflags,omitempty"`
}

// OrderInfo contains order data.
type OrderInfo struct {
	// Referral order transaction ID that created this order
	ReferralOrderTransactionId string `json:"refid,omitempty"`
	// Optional user defined reference ID
	UserReferenceId *int64 `json:"userref,omitempty"`
	// Status of order. Cf. OrderStatusEnum
	Status string `json:"status,omitempty"`
	// Unix timestamp of when order was placed.
	//
	// Unix seconds timestamp with nanoseconds as decimal part (ex: 1688666559.8974)
	OpenTimestamp string `json:"opentm,omitempty"`
	// Unix timestamp of order start time (or 0 if not set)
	//
	// Unix seconds timestamp with nanoseconds as decimal part (ex: 1688666559.8974)
	StartTimestamp string `json:"starttm,omitempty"`
	// Optional dependent on whether order type is iceberg - the visible quantity for iceberg order types
	DisplayVolume string `json:"display_volume,omitempty"`
	// Optional dependent on whether order type is iceberg - the visible quantity remaing in the order for iceberg order types
	DisplayVolumeRemain string `json:"display_volume_remain,omitempty"`
	// Unix timestamp of order end time (or 0 if not set)
	ExpireTimestamp string `json:"expiretm,omitempty"`
	// Conditional close order info (if conditional close set)
	Contingent *CloseOrderInfo `json:"contingent,omitempty"`
	// Order description info
	Description *OrderInfoDescription `json:"descr,omitempty"`
	// Unix timestamp of last change (for updates)
	//
	// Unix seconds timestamp with nanoseconds as decimal part (ex: 1688666559.8974)
	LastUpdated string `json:"lastupdated,omitempty"`
	// Volume of order (base currency)
	Volume string `json:"vol,omitempty"`
	// Volume executed (base currency)
	VolumeExecuted string `json:"vol_exec,omitempty"`
	// Total cost (quote currency unless)
	Cost string `json:"cost,omitempty"`
	// Total fee  (quote currency)
	Fee string `json:"fee,omitempty"`
	// Average price  (quote currency)
	AvgPrice string `json:"avg_price,omitempty"`
	// Stop price  (quote currency)
	StopPrice string `json:"stopprice,omitempty"`
	// Triggered limit price  (quote currency, when limit based order type triggered)
	LimitPrice string `json:"limitprice,omitempty"`
	// Comma delimited list of miscellaneous info
	Miscellaneous string `json:"misc,omitempty"`
	// Comma delimited list of order flags
	OrderFlags string `json:"oflags,omitempty"`
	// Optional - time in force.
	TimeInForce string `json:"timeinforce,omitempty"`
	// Optional - cancel reason, present for all cancellation updates (status="canceled") and for some close updates (status="closed")
	CancelReason string `json:"cancel_reason,omitempty"`
	// Optional - rate-limit counter, present if requested in subscription request.
	RateCount int `json:"ratecount,omitempty"`
}
