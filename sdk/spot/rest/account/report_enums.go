package account

// Enum for report formats
type ReportFormatEnum string

// Values for ReportFormatEnum
const (
	CSV ReportFormatEnum = "CSV"
	TSV ReportFormatEnum = "TSV"
)

// Enum for report fields. Depending on the rpeort type, only some fields are allowed:
//   - Values for trades: ordertxid, time, ordertype, price, cost, fee, vol, margin, misc, ledgers
//   - Values for ledgers: refid, time, type, aclass, asset, amount, fee, balance
type ReportFieldsEnum string

// Values for ReportFieldsEnum
const (
	//
	FieldsAll       ReportFieldsEnum = "all"
	FieldsOrderTxId ReportFieldsEnum = "ordertxid"
	FieldsTime      ReportFieldsEnum = "time"
	FieldsOrderType ReportFieldsEnum = "ordertype"
	FieldsPrice     ReportFieldsEnum = "price"
	FieldsCost      ReportFieldsEnum = "cost"
	FieldsFee       ReportFieldsEnum = "fee"
	FieldsVolume    ReportFieldsEnum = "vol"
	FieldsMargin    ReportFieldsEnum = "margin"
	FieldsMisc      ReportFieldsEnum = "misc"
	FieldsLedgers   ReportFieldsEnum = "ledgers"
	FieldsRefId     ReportFieldsEnum = "refid"
	FieldsType      ReportFieldsEnum = "type"
	FieldsAClass    ReportFieldsEnum = "aclass"
	FieldsAmount    ReportFieldsEnum = "amount"
	FieldsBalance   ReportFieldsEnum = "balance"
)

// Enum for report deletion type
type ReportDeletionEnum string

// Values for ReportDeletionEnum
const (
	DeleteReport ReportDeletionEnum = "delete"
	CancelReport ReportDeletionEnum = "cancel"
)

// Enum for report types
type ReportTypeEnum string

// Values for ReportTypeEnum
const (
	ReportTrades  ReportTypeEnum = "trades"
	ReportLedgers ReportTypeEnum = "ledgers"
)

// Enum for report export status
type ReportStatusEnum string

// Values for ReportStatusEnum
const (
	Queued     ReportStatusEnum = "Queued"
	Processing ReportStatusEnum = "Processing"
	Processed  ReportStatusEnum = "Processed"
)
