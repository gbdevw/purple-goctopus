package rest

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gbdevw/purple-goctopus/spot/rest/account"
	"github.com/gbdevw/purple-goctopus/spot/rest/common"
	"github.com/gbdevw/purple-goctopus/spot/rest/earn"
	"github.com/gbdevw/purple-goctopus/spot/rest/funding"
	"github.com/gbdevw/purple-goctopus/spot/rest/market"
	"github.com/gbdevw/purple-goctopus/spot/rest/tracing"
	"github.com/gbdevw/purple-goctopus/spot/rest/trading"
	"github.com/gbdevw/purple-goctopus/spot/rest/websocket"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// A decorator for KrakenSpotRESTClient which manages code instrumentation by using the
// OpenTelemetry framework.
type KrakenSpotRESTClientInstrumentationDecorator struct {
	// Decorated
	decorated KrakenSpotRESTClientIface
	// Tracer
	tracer trace.Tracer
}

// # Description
//
// Decorate the provided KrakenSpotRESTClientIface implementation. The function returns the decorator which
// manages tracing and code instrumentation of the decorated by using the OpenTelemetry framework.
//
// # Inputs
//
//   - decorated: The KrakenSpotRESTClientIface implentation to decorate. Must no be nil.
//   - tracerProvider: Tracer provider to use to get the tracer used by the decorator to instrument code. If nil, the global tracer provider will be used (can be a NoopTracerProvider).
//
// # Returns
//
// The decorator which decorates the provided KrakenSpotRESTClientIface implementation.
func DecorateKrakenSpotRESTClient(decorated KrakenSpotRESTClientIface, tracerProvider trace.TracerProvider) KrakenSpotRESTClientIface {
	if decorated == nil {
		// Panic if decorated is nil
		panic("decorated cannot be nil")
	}
	if tracerProvider == nil {
		// Use the global tracer provider if the provided tracer provider is nil.
		// In case the global tracer provider is not configured, its default behavior is to return a NoopTracerProvider.
		tracerProvider = otel.GetTracerProvider()
	}
	// Return decorator
	return &KrakenSpotRESTClientInstrumentationDecorator{
		decorated: decorated,
		tracer:    tracerProvider.Tracer(tracing.PackageName, trace.WithInstrumentationVersion(tracing.PackageVersion)),
	}
}

// Trace GetServerTime execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetServerTime(ctx context.Context) (*market.GetServerTimeResponse, *http.Response, error) {
	// Start a span
	ctx, span := dec.tracer.Start(ctx, tracing.TracesNamespace+".get_server_time", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetServerTime(ctx)
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetSystemStatus execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetSystemStatus(ctx context.Context) (*market.GetSystemStatusResponse, *http.Response, error) {
	// Start a span
	ctx, span := dec.tracer.Start(ctx, tracing.TracesNamespace+".get_system_status", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetSystemStatus(ctx)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttr := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttr = append(respAttr, attribute.String("status", resp.Result.Status))
		}
		span.AddEvent(tracing.TracesNamespace+".get_system_status.response", trace.WithAttributes(
			respAttr...,
		))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetAssetInfo execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetAssetInfo(ctx context.Context, opts *market.GetAssetInfoRequestOptions) (*market.GetAssetInfoResponse, *http.Response, error) {
	// Build attributes which record request settings
	reqAttr := []attribute.KeyValue{}
	if opts != nil {
		if len(opts.Assets) > 0 {
			reqAttr = append(reqAttr, attribute.StringSlice("assets", opts.Assets))
		}
		if opts.AssetClass != "" {
			reqAttr = append(reqAttr, attribute.String("aclass", opts.AssetClass))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_asset_info",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttr...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetAssetInfo(ctx, opts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttr := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttr = append(respAttr, attribute.Int("count", len(resp.Result)))
		}
		span.AddEvent(tracing.TracesNamespace+".get_asset_info.response", trace.WithAttributes(
			respAttr...,
		))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetTradableAssetPairs execution
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetTradableAssetPairs(ctx context.Context, opts *market.GetTradableAssetPairsRequestOptions) (*market.GetTradableAssetPairsResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{}
	if opts != nil {
		if len(opts.Pairs) > 0 {
			reqAttributes = append(reqAttributes, attribute.StringSlice("pairs", opts.Pairs))
		}
		if opts.Info != "" {
			reqAttributes = append(reqAttributes, attribute.String("info", opts.Info))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_tradable_asset_pairs",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetTradableAssetPairs(ctx, opts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		span.AddEvent(tracing.TracesNamespace+".get_tradable_asset_pairs.response", trace.WithAttributes(
			attribute.StringSlice("error", resp.Error),
			attribute.Int("count", len(resp.Result)),
		))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetTickerInformation execution
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetTickerInformation(ctx context.Context, opts *market.GetTickerInformationRequestOptions) (*market.GetTickerInformationResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{}
	if opts != nil {
		if len(opts.Pairs) > 0 {
			reqAttributes = append(reqAttributes, attribute.StringSlice("pairs", opts.Pairs))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_ticker_information",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetTickerInformation(ctx, opts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		span.AddEvent(tracing.TracesNamespace+".get_ticker_information.response", trace.WithAttributes(
			attribute.StringSlice("error", resp.Error),
			attribute.Int("count", len(resp.Result)),
		))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetOHLCData execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetOHLCData(ctx context.Context, params market.GetOHLCDataRequestParameters, opts *market.GetOHLCDataRequestOptions) (*market.GetOHLCDataResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{attribute.String("pair", params.Pair)}
	if opts != nil {
		if opts.Interval != 0 {
			reqAttributes = append(reqAttributes, attribute.Int64("interval", opts.Interval))
		}
		if opts.Since != 0 {
			reqAttributes = append(reqAttributes, attribute.Int64("since", opts.Since))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_ohlc_data",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetOHLCData(ctx, params, opts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttributes = append(
				respAttributes,
				attribute.String("pair", resp.Result.PairId),
				attribute.Int("count", len(resp.Result.Data)),
				attribute.Int64("last", resp.Result.Last))
		}
		span.AddEvent(tracing.TracesNamespace+".get_ohlc_data.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetOrderBook execute
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetOrderBook(ctx context.Context, params market.GetOrderBookRequestParameters, opts *market.GetOrderBookRequestOptions) (*market.GetOrderBookResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{attribute.String("pair", params.Pair)}
	if opts != nil {
		if opts.Count != 0 {
			reqAttributes = append(reqAttributes, attribute.Int("count", opts.Count))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_order_book",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetOrderBook(ctx, params, opts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttributes = append(
				respAttributes,
				attribute.String("pair", resp.Result.PairId),
				attribute.Int("asks.count", len(resp.Result.Asks)),
				attribute.Int("bids.count", len(resp.Result.Bids)))
		}
		span.AddEvent(tracing.TracesNamespace+".get_order_book.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetRecentTrades execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetRecentTrades(ctx context.Context, params market.GetRecentTradesRequestParameters, opts *market.GetRecentTradesRequestOptions) (*market.GetRecentTradesResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{attribute.String("pair", params.Pair)}
	if opts != nil {
		if opts.Count != 0 {
			reqAttributes = append(reqAttributes, attribute.Int("count", opts.Count))
		}
		if opts.Since != 0 {
			reqAttributes = append(reqAttributes, attribute.Int64("since", opts.Since))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_recent_trades",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetRecentTrades(ctx, params, opts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttributes = append(
				respAttributes,
				attribute.String("pair", resp.Result.PairId),
				attribute.Int("count", len(resp.Result.Trades)),
				attribute.Int64("last", resp.Result.Last))
		}
		span.AddEvent(tracing.TracesNamespace+".get_recent_trades.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetRecentSpreads execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetRecentSpreads(ctx context.Context, params market.GetRecentSpreadsRequestParameters, opts *market.GetRecentSpreadsRequestOptions) (*market.GetRecentSpreadsResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{attribute.String("pair", params.Pair)}
	if opts != nil {
		if opts.Since != 0 {
			reqAttributes = append(reqAttributes, attribute.Int64("since", opts.Since))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_recent_spreads",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetRecentSpreads(ctx, params, opts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttributes = append(
				respAttributes,
				attribute.String("pair", resp.Result.PairId),
				attribute.Int("count", len(resp.Result.Spreads)),
				attribute.Int64("last", resp.Result.Last))
		}
		span.AddEvent(tracing.TracesNamespace+".get_recent_spreads.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetAccountBalance execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetAccountBalance(ctx context.Context, nonce int64, secopts *common.SecurityOptions) (*account.GetAccountBalanceResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{attribute.Int64("nonce", nonce)}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_account_balance",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetAccountBalance(ctx, nonce, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{
			attribute.StringSlice("error", resp.Error),
			attribute.Int("count", len(resp.Result)),
		}
		span.AddEvent(tracing.TracesNamespace+".get_account_balance.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetExtendedBalance execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetExtendedBalance(ctx context.Context, nonce int64, secopts *common.SecurityOptions) (*account.GetExtendedBalanceResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{attribute.Int64("nonce", nonce)}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_extended_balance",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetExtendedBalance(ctx, nonce, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{
			attribute.StringSlice("error", resp.Error),
			attribute.Int("count", len(resp.Result)),
		}
		span.AddEvent(tracing.TracesNamespace+".get_extended_balance.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetTradeBalance execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetTradeBalance(ctx context.Context, nonce int64, opts *account.GetTradeBalanceRequestOptions, secopts *common.SecurityOptions) (*account.GetTradeBalanceResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{attribute.Int64("nonce", nonce)}
	if opts != nil {
		if opts.Asset != "" {
			reqAttributes = append(reqAttributes, attribute.String("asset", opts.Asset))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_trade_balance",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetTradeBalance(ctx, nonce, opts, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{
			attribute.StringSlice("error", resp.Error),
		}
		span.AddEvent(tracing.TracesNamespace+".get_trade_balance.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetOpenOrders execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetOpenOrders(ctx context.Context, nonce int64, opts *account.GetOpenOrdersRequestOptions, secopts *common.SecurityOptions) (*account.GetOpenOrdersResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{attribute.Int64("nonce", nonce)}
	if opts != nil {
		reqAttributes = append(reqAttributes, attribute.Bool("trades", opts.Trades))
		if opts.UserReference != nil {
			reqAttributes = append(reqAttributes, attribute.Int64("userref", *opts.UserReference))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_open_orders",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetOpenOrders(ctx, nonce, opts, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttributes = append(respAttributes, attribute.Int("count", len(resp.Result.Open)))
		}
		span.AddEvent(tracing.TracesNamespace+".get_open_orders.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetClosedOrders execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetClosedOrders(ctx context.Context, nonce int64, opts *account.GetClosedOrdersRequestOptions, secopts *common.SecurityOptions) (*account.GetClosedOrdersResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{attribute.Int64("nonce", nonce)}
	if opts != nil {
		reqAttributes = append(reqAttributes, attribute.Bool("trades", opts.Trades))
		reqAttributes = append(reqAttributes, attribute.Bool("consolidate_taker", opts.ConsolidateTaker))
		if opts.UserReference != nil {
			reqAttributes = append(reqAttributes, attribute.Int64("userref", *opts.UserReference))
		}
		if opts.Closetime != "" {
			reqAttributes = append(reqAttributes, attribute.String("closetime", opts.Closetime))
		}
		if opts.Start != "" {
			reqAttributes = append(reqAttributes, attribute.String("start", opts.Start))
		}
		if opts.End != "" {
			reqAttributes = append(reqAttributes, attribute.String("end", opts.End))
		}
		if opts.Offset != 0 {
			reqAttributes = append(reqAttributes, attribute.Int64("ofs", opts.Offset))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_closed_orders",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetClosedOrders(ctx, nonce, opts, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttributes = append(respAttributes, attribute.Int("count", len(resp.Result.Closed)))
		}
		span.AddEvent(tracing.TracesNamespace+".get_closed_orders.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace QueryOrdersInfo execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) QueryOrdersInfo(ctx context.Context, nonce int64, params account.QueryOrdersInfoParameters, opts *account.QueryOrdersInfoRequestOptions, secopts *common.SecurityOptions) (*account.QueryOrdersInfoResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{
		attribute.Int64("nonce", nonce),
		attribute.StringSlice("txid", params.TxId),
	}
	if opts != nil {
		reqAttributes = append(reqAttributes, attribute.Bool("trades", opts.Trades))

		if opts.UserReference != nil {
			reqAttributes = append(reqAttributes, attribute.Int64("userref", *opts.UserReference))
		}
		if opts.ConsolidateTaker != nil {
			reqAttributes = append(reqAttributes, attribute.Bool("consolidate_taker", *opts.ConsolidateTaker))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".query_orders_info",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.QueryOrdersInfo(ctx, nonce, params, opts, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{
			attribute.StringSlice("error", resp.Error),
			attribute.Int("count", len(resp.Result)),
		}
		span.AddEvent(tracing.TracesNamespace+".query_orders_info.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetTradesHistory execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetTradesHistory(ctx context.Context, nonce int64, opts *account.GetTradesHistoryRequestOptions, secopts *common.SecurityOptions) (*account.GetTradesHistoryResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{attribute.Int64("nonce", nonce)}
	if opts != nil {
		reqAttributes = append(reqAttributes, attribute.Bool("trades", opts.Trades))
		reqAttributes = append(reqAttributes, attribute.Bool("consolidate_taker", opts.ConsolidateTaker))
		if opts.Start != "" {
			reqAttributes = append(reqAttributes, attribute.String("start", opts.Start))
		}
		if opts.End != "" {
			reqAttributes = append(reqAttributes, attribute.String("end", opts.End))
		}
		if opts.Offset != 0 {
			reqAttributes = append(reqAttributes, attribute.Int64("ofs", opts.Offset))
		}
		if opts.Type != "" {
			reqAttributes = append(reqAttributes, attribute.String("type", opts.Type))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_trades_history",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetTradesHistory(ctx, nonce, opts, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttributes = append(respAttributes, attribute.Int("count", len(resp.Result.Trades)))
		}
		span.AddEvent(tracing.TracesNamespace+".get_trades_history.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace QueryTradesInfo execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) QueryTradesInfo(ctx context.Context, nonce int64, params account.QueryTradesRequestParameters, opts *account.QueryTradesRequestOptions, secopts *common.SecurityOptions) (*account.QueryTradesInfoResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{
		attribute.Int64("nonce", nonce),
		attribute.StringSlice("txid", params.TransactionIds),
	}
	if opts != nil {
		reqAttributes = append(reqAttributes, attribute.Bool("trades", opts.Trades))
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".query_trades_info",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.QueryTradesInfo(ctx, nonce, params, opts, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{
			attribute.StringSlice("error", resp.Error),
			attribute.Int("count", len(resp.Result)),
		}
		span.AddEvent(tracing.TracesNamespace+".query_trades_info.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetOpenPositions execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetOpenPositions(ctx context.Context, nonce int64, opts *account.GetOpenPositionsRequestOptions, secopts *common.SecurityOptions) (*account.GetOpenPositionsResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{attribute.Int64("nonce", nonce)}
	if opts != nil {
		reqAttributes = append(reqAttributes, attribute.Bool("docalcs", opts.DoCalcs))
		if len(opts.TransactionIds) > 0 {
			reqAttributes = append(reqAttributes, attribute.StringSlice("txid", opts.TransactionIds))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_open_positions",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetOpenPositions(ctx, nonce, opts, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{
			attribute.StringSlice("error", resp.Error),
			attribute.Int("count", len(resp.Result)),
		}
		span.AddEvent(tracing.TracesNamespace+".get_open_positions.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetLedgersInfo execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetLedgersInfo(ctx context.Context, nonce int64, opts *account.GetLedgersInfoRequestOptions, secopts *common.SecurityOptions) (*account.GetLedgersInfoResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{attribute.Int64("nonce", nonce)}
	if opts != nil {
		reqAttributes = append(reqAttributes, attribute.Bool("without_count", opts.WithoutCount))
		if len(opts.Assets) > 0 {
			reqAttributes = append(reqAttributes, attribute.StringSlice("asset", opts.Assets))
		}
		if opts.AssetClass != "" {
			reqAttributes = append(reqAttributes, attribute.String("aclass", opts.AssetClass))
		}
		if opts.End != "" {
			reqAttributes = append(reqAttributes, attribute.String("end", opts.End))
		}
		if opts.Start != "" {
			reqAttributes = append(reqAttributes, attribute.String("start", opts.Start))
		}
		if opts.Type != "" {
			reqAttributes = append(reqAttributes, attribute.String("type", opts.Type))
		}
		if opts.Offset != 0 {
			reqAttributes = append(reqAttributes, attribute.Int64("ofs", opts.Offset))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_ledgers_info",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetLedgersInfo(ctx, nonce, opts, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttributes = append(respAttributes, attribute.Int("count", len(resp.Result.Ledgers)))
		}
		span.AddEvent(tracing.TracesNamespace+".get_ledgers_info.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace QueryLedgers execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) QueryLedgers(ctx context.Context, nonce int64, params account.QueryLedgersRequestParameters, opts *account.QueryLedgersRequestOptions, secopts *common.SecurityOptions) (*account.QueryLedgersResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{
		attribute.Int64("nonce", nonce),
		attribute.StringSlice("id", params.Id),
	}
	if opts != nil {
		reqAttributes = append(reqAttributes, attribute.Bool("trades", opts.Trades))
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".query_ledgers",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.QueryLedgers(ctx, nonce, params, opts, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{
			attribute.StringSlice("error", resp.Error),
			attribute.Int("count", len(resp.Result)),
		}
		span.AddEvent(tracing.TracesNamespace+".query_ledgers.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetTradeVolume execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetTradeVolume(ctx context.Context, nonce int64, opts *account.GetTradeVolumeRequestOptions, secopts *common.SecurityOptions) (*account.GetTradeVolumeResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{attribute.Int64("nonce", nonce)}
	if opts != nil {
		reqAttributes = append(reqAttributes, attribute.StringSlice("pair", opts.Pairs))
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_trade_volume",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetTradeVolume(ctx, nonce, opts, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{
			attribute.StringSlice("error", resp.Error),
		}
		span.AddEvent(tracing.TracesNamespace+".get_trade_volume.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace RequestExportReport execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) RequestExportReport(ctx context.Context, nonce int64, params account.RequestExportReportRequestParameters, opts *account.RequestExportReportRequestOptions, secopts *common.SecurityOptions) (*account.RequestExportReportResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{
		attribute.Int64("nonce", nonce),
		attribute.String("report", params.Report),
		attribute.String("description", params.Description),
	}
	if opts != nil {
		if len(opts.Fields) > 0 {
			reqAttributes = append(reqAttributes, attribute.StringSlice("fields", opts.Fields))
		}
		if opts.Format != "" {
			reqAttributes = append(reqAttributes, attribute.String("format", opts.Format))
		}
		if opts.EndTm != 0 {
			reqAttributes = append(reqAttributes, attribute.Int64("endtm", opts.EndTm))
		}
		if opts.StartTm != 0 {
			reqAttributes = append(reqAttributes, attribute.Int64("starttm", opts.StartTm))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".request_export_report",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.RequestExportReport(ctx, nonce, params, opts, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttributes = append(respAttributes, attribute.String("id", resp.Result.Id))
		}
		span.AddEvent(tracing.TracesNamespace+".request_export_report.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetExportReportStatus execution
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetExportReportStatus(ctx context.Context, nonce int64, params account.GetExportReportStatusRequestParameters, secopts *common.SecurityOptions) (*account.GetExportReportStatusResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{
		attribute.Int64("nonce", nonce),
		attribute.String("report", params.Report),
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_export_report_status",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetExportReportStatus(ctx, nonce, params, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{
			attribute.StringSlice("error", resp.Error),
			attribute.Int("count", len(resp.Result)),
		}
		span.AddEvent(tracing.TracesNamespace+".get_export_report_status.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace RetrieveDataExport execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) RetrieveDataExport(ctx context.Context, nonce int64, params account.RetrieveDataExportParameters, secopts *common.SecurityOptions) (*account.RetrieveDataExportResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{
		attribute.Int64("nonce", nonce),
		attribute.String("id", params.Id),
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".retrieve_data_export",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.RetrieveDataExport(ctx, nonce, params, secopts)
	// Trace error and set span status
	tracing.TraceErrorAndSetStatus(span, err)
	// Return results
	return resp, httpresp, err
}

// Trace DeleteExportReport execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) DeleteExportReport(ctx context.Context, nonce int64, params account.DeleteExportReportRequestParameters, secopts *common.SecurityOptions) (*account.DeleteExportReportResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{
		attribute.Int64("nonce", nonce),
		attribute.String("id", params.Id),
		attribute.String("type", params.Type),
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".delete_export_report",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.DeleteExportReport(ctx, nonce, params, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttributes = append(
				respAttributes,
				attribute.Bool("cancel", resp.Result.Cancel),
				attribute.Bool("cancel", resp.Result.Delete))
		}
		span.AddEvent(tracing.TracesNamespace+".delete_export_report.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace AddOrder execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) AddOrder(ctx context.Context, nonce int64, params trading.AddOrderRequestParameters, opts *trading.AddOrderRequestOptions, secopts *common.SecurityOptions) (*trading.AddOrderResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{
		attribute.Int64("nonce", nonce),
		attribute.String("pair", params.Pair),
		attribute.String("ordertype", params.Order.OrderType),
		attribute.String("type", params.Order.Type),
		attribute.String("volume", params.Order.Volume),
		attribute.Bool("reduce_only", params.Order.ReduceOnly),
	}
	if params.Order.DisplayedVolume != "" {
		reqAttributes = append(reqAttributes, attribute.String("displayvol", params.Order.DisplayedVolume))
	}
	if params.Order.Price != "" {
		reqAttributes = append(reqAttributes, attribute.String("price", params.Order.Price))
	}
	if params.Order.Price2 != "" {
		reqAttributes = append(reqAttributes, attribute.String("price2", params.Order.Price2))
	}
	if params.Order.Trigger != "" {
		reqAttributes = append(reqAttributes, attribute.String("trigger", params.Order.Trigger))
	}
	if params.Order.Leverage != "" {
		reqAttributes = append(reqAttributes, attribute.String("leverage", params.Order.Leverage))
	}
	if params.Order.StpType != "" {
		reqAttributes = append(reqAttributes, attribute.String("stptype", params.Order.StpType))
	}
	if params.Order.OrderFlags != "" {
		reqAttributes = append(reqAttributes, attribute.String("oflags", params.Order.OrderFlags))
	}
	if params.Order.TimeInForce != "" {
		reqAttributes = append(reqAttributes, attribute.String("timeinforce", params.Order.TimeInForce))
	}
	if params.Order.ScheduledStartTime != "" {
		reqAttributes = append(reqAttributes, attribute.String("starttm", params.Order.ScheduledStartTime))
	}
	if params.Order.ExpirationTime != "" {
		reqAttributes = append(reqAttributes, attribute.String("expiretm", params.Order.ExpirationTime))
	}
	if params.Order.Close.OrderType != "" {
		reqAttributes = append(reqAttributes, attribute.String("close[ordertype]", params.Order.Close.OrderType))
	}
	if params.Order.Close.Price != "" {
		reqAttributes = append(reqAttributes, attribute.String("close[price]", params.Order.Close.Price))
	}
	if params.Order.Close.Price2 != "" {
		reqAttributes = append(reqAttributes, attribute.String("close[price2]", params.Order.Close.Price2))
	}
	if opts != nil {
		reqAttributes = append(reqAttributes, attribute.Bool("validate", opts.Validate))
		if !opts.Deadline.IsZero() {
			reqAttributes = append(reqAttributes, attribute.String("deadline", opts.Deadline.Format(time.RFC3339)))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".add_order",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.AddOrder(ctx, nonce, params, opts, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttributes = append(
				respAttributes,
				attribute.StringSlice("txid", resp.Result.TransactionIDs),
				attribute.String("description", resp.Result.Description.Order),
				attribute.String("close", resp.Result.Description.Close))
		}
		span.AddEvent(tracing.TracesNamespace+".add_order.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace AddOrderBatch execution
func (dec *KrakenSpotRESTClientInstrumentationDecorator) AddOrderBatch(ctx context.Context, nonce int64, params trading.AddOrderBatchRequestParameters, opts *trading.AddOrderBatchOptions, secopts *common.SecurityOptions) (*trading.AddOrderBatchResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{
		attribute.Int64("nonce", nonce),
		attribute.String("pair", params.Pair),
	}
	for index, order := range params.Orders {
		reqAttributes = append(reqAttributes,
			attribute.String(fmt.Sprintf("orders[%d][%s]", index, "ordertype"), order.OrderType),
			attribute.String(fmt.Sprintf("orders[%d][%s]", index, "type"), order.Type),
			attribute.String(fmt.Sprintf("orders[%d][%s]", index, "volume"), order.Volume),
			attribute.Bool(fmt.Sprintf("orders[%d][%s]", index, "reduce_only"), order.ReduceOnly))
		if order.DisplayedVolume != "" {
			reqAttributes = append(reqAttributes, attribute.String(fmt.Sprintf("orders[%d][%s]", index, "displayvol"), order.DisplayedVolume))
		}
		if order.Price != "" {
			reqAttributes = append(reqAttributes, attribute.String(fmt.Sprintf("orders[%d][%s]", index, "price"), order.Price))
		}
		if order.Price2 != "" {
			reqAttributes = append(reqAttributes, attribute.String(fmt.Sprintf("orders[%d][%s]", index, "price2"), order.Price2))
		}
		if order.Trigger != "" {
			reqAttributes = append(reqAttributes, attribute.String(fmt.Sprintf("orders[%d][%s]", index, "trigger"), order.Trigger))
		}
		if order.Leverage != "" {
			reqAttributes = append(reqAttributes, attribute.String(fmt.Sprintf("orders[%d][%s]", index, "leverage"), order.Leverage))
		}
		if order.StpType != "" {
			reqAttributes = append(reqAttributes, attribute.String(fmt.Sprintf("orders[%d][%s]", index, "stptype"), order.StpType))
		}
		if order.OrderFlags != "" {
			reqAttributes = append(reqAttributes, attribute.String(fmt.Sprintf("orders[%d][%s]", index, "oflags"), order.OrderFlags))
		}
		if order.TimeInForce != "" {
			reqAttributes = append(reqAttributes, attribute.String(fmt.Sprintf("orders[%d][%s]", index, "timeinforce"), order.TimeInForce))
		}
		if order.ScheduledStartTime != "" {
			reqAttributes = append(reqAttributes, attribute.String(fmt.Sprintf("orders[%d][%s]", index, "starttm"), order.ScheduledStartTime))
		}
		if order.ExpirationTime != "" {
			reqAttributes = append(reqAttributes, attribute.String(fmt.Sprintf("orders[%d][%s]", index, "expiretm"), order.ExpirationTime))
		}
		if order.Close.OrderType != "" {
			reqAttributes = append(reqAttributes, attribute.String(fmt.Sprintf("orders[%d][%s][%s]", index, "close", "ordertype"), order.Close.OrderType))
		}
		if order.Close.Price != "" {
			reqAttributes = append(reqAttributes, attribute.String(fmt.Sprintf("orders[%d][%s][%s]", index, "close", "price"), order.Close.Price))
		}
		if order.Close.Price2 != "" {
			reqAttributes = append(reqAttributes, attribute.String(fmt.Sprintf("orders[%d][%s][%s]", index, "close", "price2"), order.Close.Price2))
		}
	}
	if opts != nil {
		reqAttributes = append(reqAttributes, attribute.Bool("validate", opts.Validate))
		if !opts.Deadline.IsZero() {
			reqAttributes = append(reqAttributes, attribute.String("deadline", opts.Deadline.Format(time.RFC3339)))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".add_order_batch",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.AddOrderBatch(ctx, nonce, params, opts, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttributes = append(respAttributes, attribute.Int("count", len(resp.Result.Orders)))
			for index, order := range resp.Result.Orders {
				respAttributes = append(
					respAttributes,
					attribute.String(fmt.Sprintf("orders[%d][%s]", index, "txid"), order.Id),
					attribute.String(fmt.Sprintf("orders[%d][%s]", index, "description"), order.Description.Order),
					attribute.String(fmt.Sprintf("orders[%d][%s]", index, "close"), order.Description.Close),
				)
			}
		}
		span.AddEvent(tracing.TracesNamespace+".add_order_batch.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace EditOrder execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) EditOrder(ctx context.Context, nonce int64, params trading.EditOrderRequestParameters, opts *trading.EditOrderRequestOptions, secopts *common.SecurityOptions) (*trading.EditOrderResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{
		attribute.Int64("nonce", nonce),
		attribute.String("pair", params.Pair),
		attribute.String("txid", params.Id),
	}
	if opts != nil {
		if opts.NewUserReference != "" {
			reqAttributes = append(reqAttributes, attribute.String("userref", opts.NewUserReference))
		}
		if opts.NewVolume != "" {
			reqAttributes = append(reqAttributes, attribute.String("volume", opts.NewVolume))
		}
		if opts.NewDisplayedVolume != "" {
			reqAttributes = append(reqAttributes, attribute.String("displayvol", opts.NewDisplayedVolume))
		}
		if opts.Price != "" {
			reqAttributes = append(reqAttributes, attribute.String("price", opts.Price))
		}
		if opts.Price2 != "" {
			reqAttributes = append(reqAttributes, attribute.String("price2", opts.Price2))
		}
		if len(opts.OFlags) > 0 {
			reqAttributes = append(reqAttributes, attribute.StringSlice("oflags", opts.OFlags))
		}
		reqAttributes = append(reqAttributes, attribute.Bool("validate", opts.Validate))
		reqAttributes = append(reqAttributes, attribute.Bool("cancel_response", opts.CancelResponse))
		if !opts.Deadline.IsZero() {
			reqAttributes = append(reqAttributes, attribute.String("deadline", opts.Deadline.Format(time.RFC3339)))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".edit_order",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.EditOrder(ctx, nonce, params, opts, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			if resp.Result.Description != nil {
				respAttributes = append(
					respAttributes,
					attribute.String("description", resp.Result.Description.Order),
					attribute.String("close", resp.Result.Description.Close))
			}
			if resp.Result.TransactionID != "" {
				respAttributes = append(respAttributes, attribute.String("txid", resp.Result.TransactionID))
			}
			if resp.Result.NewUserReference != nil {
				respAttributes = append(respAttributes, attribute.Int64("newuserref", *resp.Result.NewUserReference))
			}
			if resp.Result.OldUserReference != nil {
				respAttributes = append(respAttributes, attribute.Int64("txid", *resp.Result.OldUserReference))
			}
			if resp.Result.OrdersCancelled != 0 {
				respAttributes = append(respAttributes, attribute.Int("orders_cancelled", resp.Result.OrdersCancelled))
			}
			if resp.Result.OriginalTransactionID != "" {
				respAttributes = append(respAttributes, attribute.String("originaltxid", resp.Result.OriginalTransactionID))
			}
			if resp.Result.Status != "" {
				respAttributes = append(respAttributes, attribute.String("status", resp.Result.Status))
			}
			if resp.Result.Volume != "" {
				respAttributes = append(respAttributes, attribute.String("volume", resp.Result.Volume))
			}
			if resp.Result.Price != "" {
				respAttributes = append(respAttributes, attribute.String("price", resp.Result.Price))
			}
			if resp.Result.Price2 != "" {
				respAttributes = append(respAttributes, attribute.String("price2", resp.Result.Price2))
			}
			if resp.Result.ErrorMsg != "" {
				respAttributes = append(respAttributes, attribute.String("error_message", resp.Result.ErrorMsg))
			}
		}
		span.AddEvent(tracing.TracesNamespace+".edit_order.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace CancelOrder execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) CancelOrder(ctx context.Context, nonce int64, params trading.CancelOrderRequestParameters, secopts *common.SecurityOptions) (*trading.CancelOrderResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{
		attribute.Int64("nonce", nonce),
		attribute.String("txid", params.Id),
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".cancel_order",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.CancelOrder(ctx, nonce, params, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttributes = append(respAttributes, attribute.Int("count", resp.Result.Count))
		}
		span.AddEvent(tracing.TracesNamespace+".cancel_order.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace CancelAllOrders execution
func (dec *KrakenSpotRESTClientInstrumentationDecorator) CancelAllOrders(ctx context.Context, nonce int64, secopts *common.SecurityOptions) (*trading.CancelAllOrdersResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{attribute.Int64("nonce", nonce)}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".cancel_all_orders",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.CancelAllOrders(ctx, nonce, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttributes = append(respAttributes, attribute.Int("count", resp.Result.Count))
		}
		span.AddEvent(tracing.TracesNamespace+".cancel_all_orders.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace CancelAllOrdersAfterX execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) CancelAllOrdersAfterX(ctx context.Context, nonce int64, params trading.CancelAllOrdersAfterXRequestParameters, secopts *common.SecurityOptions) (*trading.CancelAllOrdersAfterXResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{
		attribute.Int64("nonce", nonce),
		attribute.Int64("timeout", params.Timeout),
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".cancel_all_orders_after_x",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.CancelAllOrdersAfterX(ctx, nonce, params, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttributes = append(
				respAttributes,
				attribute.String("currentTime", resp.Result.CurrentTime.Format(time.RFC3339)),
				attribute.String("triggerTime", resp.Result.TriggerTime.Format(time.RFC3339)))
		}
		span.AddEvent(tracing.TracesNamespace+".cancel_all_orders_after_x.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace CancelOrderBatch execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) CancelOrderBatch(ctx context.Context, nonce int64, params trading.CancelOrderBatchRequestParameters, secopts *common.SecurityOptions) (*trading.CancelOrderBatchResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{
		attribute.Int64("nonce", nonce),
		attribute.StringSlice("orders", params.OrderIds),
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".cancel_order_batch",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.CancelOrderBatch(ctx, nonce, params, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttributes = append(respAttributes, attribute.Int("count", resp.Result.Count))
		}
		span.AddEvent(tracing.TracesNamespace+".cancel_order_batch.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetDepositMethods - Retrieve methods available for depositing a particular asset.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - params: CancelOrderBatch request parameters.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - GetDepositMethodsResponse: The parsed response from Kraken API.
//   - http.Response: A reference to the raw HTTP response received from Kraken API.
//   - error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
//
// # Note on error
//
// The error is set only when something wrong has happened either at the HTTP level (while building the request,
// when the server is unreachable, when the API replies with a status code different from 200, ...) , when
// an error happens while parsing the response JSON payload (in that case, error is json.UnmarshalTypeError) or
// when context has expired.
//
// An nil error does not mean everything is OK: You also have to check the response error field for specific
// errors from Kraken API.
//
// # Note on the http.Response
//
// A reference to the received http.Response is always returned but it may be nil if no response was received.
// Some endpoints of the Kraken API include tracing metadata in the response headers. The reference can be used
// to extract the metadata (or any other kind of data that are not used by the API client directly).
//
// Please note response body will always be closed except for RetrieveDataExport.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetDepositMethods(ctx context.Context, nonce int64, params funding.GetDepositMethodsRequestParameters, secopts *common.SecurityOptions) (*funding.GetDepositMethodsResponse, *http.Response, error)

// Trace GetDepositAddresses - Retrieve (or generate a new) deposit addresses for a particular asset and method.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - params: GetDepositAddresses request parameters.
//   - opts: GetDepositAddresses request options. A nil value triggers all default behaviors.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - GetDepositAddressesResponse: The parsed response from Kraken API.
//   - http.Response: A reference to the raw HTTP response received from Kraken API.
//   - error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
//
// # Note on error
//
// The error is set only when something wrong has happened either at the HTTP level (while building the request,
// when the server is unreachable, when the API replies with a status code different from 200, ...) , when
// an error happens while parsing the response JSON payload (in that case, error is json.UnmarshalTypeError) or
// when context has expired.
//
// An nil error does not mean everything is OK: You also have to check the response error field for specific
// errors from Kraken API.
//
// # Note on the http.Response
//
// A reference to the received http.Response is always returned but it may be nil if no response was received.
// Some endpoints of the Kraken API include tracing metadata in the response headers. The reference can be used
// to extract the metadata (or any other kind of data that are not used by the API client directly).
//
// Please note response body will always be closed except for RetrieveDataExport.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetDepositAddresses(ctx context.Context, nonce int64, params funding.GetDepositAddressesRequestParameters, opts *funding.GetDepositAddressesRequestOptions, secopts *common.SecurityOptions) (*funding.GetDepositAddressesResponse, *http.Response, error)

// Trace GetStatusOfRecentDeposits - Retrieve information about recent deposits. Results are sorted
// by recency, use the cursor parameter to iterate through list of deposits (page size equal
// to value of limit) from newest to oldest.
//
// Please note pagination usage is forced as the response format is too different.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - opts: GetStatusOfRecentDeposits request options. A nil value triggers all default behaviors.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - GetStatusOfRecentDepositsResponse: The parsed response from Kraken API.
//   - http.Response: A reference to the raw HTTP response received from Kraken API.
//   - error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
//
// # Note on error
//
// The error is set only when something wrong has happened either at the HTTP level (while building the request,
// when the server is unreachable, when the API replies with a status code different from 200, ...) , when
// an error happens while parsing the response JSON payload (in that case, error is json.UnmarshalTypeError) or
// when context has expired.
//
// An nil error does not mean everything is OK: You also have to check the response error field for specific
// errors from Kraken API.
//
// # Note on the http.Response
//
// A reference to the received http.Response is always returned but it may be nil if no response was received.
// Some endpoints of the Kraken API include tracing metadata in the response headers. The reference can be used
// to extract the metadata (or any other kind of data that are not used by the API client directly).
//
// Please note response body will always be closed except for RetrieveDataExport.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetStatusOfRecentDeposits(ctx context.Context, nonce int64, opts *funding.GetStatusOfRecentDepositsRequestOptions, secopts *common.SecurityOptions) (*funding.GetStatusOfRecentDepositsResponse, *http.Response, error)

// Trace GetWithdrawalMethods - Retrieve a list of withdrawal addresses available for the user.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - opts: GetWithdrawalMethods request options.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - GetWithdrawalMethodsResponse: The parsed response from Kraken API.
//   - http.Response: A reference to the raw HTTP response received from Kraken API.
//   - error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
//
// # Note on error
//
// The error is set only when something wrong has happened either at the HTTP level (while building the request,
// when the server is unreachable, when the API replies with a status code different from 200, ...) , when
// an error happens while parsing the response JSON payload (in that case, error is json.UnmarshalTypeError) or
// when context has expired.
//
// An nil error does not mean everything is OK: You also have to check the response error field for specific
// errors from Kraken API.
//
// # Note on the http.Response
//
// A reference to the received http.Response is always returned but it may be nil if no response was received.
// Some endpoints of the Kraken API include tracing metadata in the response headers. The reference can be used
// to extract the metadata (or any other kind of data that are not used by the API client directly).
//
// Please note response body will always be closed except for RetrieveDataExport.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetWithdrawalMethods(ctx context.Context, nonce int64, opts *funding.GetWithdrawalMethodsRequestOptions, secopts *common.SecurityOptions) (*funding.GetWithdrawalMethodsResponse, *http.Response, error)

// Trace GetWithdrawalAddresses - Retrieve a list of withdrawal addresses available for the user.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - opts: GetWithdrawalAddresses request options.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - GetWithdrawalAddressesResponse: The parsed response from Kraken API.
//   - http.Response: A reference to the raw HTTP response received from Kraken API.
//   - error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
//
// # Note on error
//
// The error is set only when something wrong has happened either at the HTTP level (while building the request,
// when the server is unreachable, when the API replies with a status code different from 200, ...) , when
// an error happens while parsing the response JSON payload (in that case, error is json.UnmarshalTypeError) or
// when context has expired.
//
// An nil error does not mean everything is OK: You also have to check the response error field for specific
// errors from Kraken API.
//
// # Note on the http.Response
//
// A reference to the received http.Response is always returned but it may be nil if no response was received.
// Some endpoints of the Kraken API include tracing metadata in the response headers. The reference can be used
// to extract the metadata (or any other kind of data that are not used by the API client directly).
//
// Please note response body will always be closed except for RetrieveDataExport.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetWithdrawalAddresses(ctx context.Context, nonce int64, opts *funding.GetWithdrawalAddressesRequestOptions, secopts *common.SecurityOptions) (*funding.GetWithdrawalAddressesResponse, *http.Response, error)

// Trace GetWithdrawalInformation - Retrieve a list of withdrawal methods available for the user.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - params: GetWithdrawalInformation request parameters.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - GetWithdrawalInformationResponse: The parsed response from Kraken API.
//   - http.Response: A reference to the raw HTTP response received from Kraken API.
//   - error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
//
// # Note on error
//
// The error is set only when something wrong has happened either at the HTTP level (while building the request,
// when the server is unreachable, when the API replies with a status code different from 200, ...) , when
// an error happens while parsing the response JSON payload (in that case, error is json.UnmarshalTypeError) or
// when context has expired.
//
// An nil error does not mean everything is OK: You also have to check the response error field for specific
// errors from Kraken API.
//
// # Note on the http.Response
//
// A reference to the received http.Response is always returned but it may be nil if no response was received.
// Some endpoints of the Kraken API include tracing metadata in the response headers. The reference can be used
// to extract the metadata (or any other kind of data that are not used by the API client directly).
//
// Please note response body will always be closed except for RetrieveDataExport.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetWithdrawalInformation(ctx context.Context, nonce int64, params funding.GetWithdrawalInformationRequestParameters, secopts *common.SecurityOptions) (*funding.GetWithdrawalInformationResponse, *http.Response, error)

// Trace WithdrawFunds - Make a withdrawal request.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - params: WithdrawFunds request parameters.
//   - opts: WithdrawFunds request options. A nil value triggers all default behaviors.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - WithdrawFundsResponse: The parsed response from Kraken API.
//   - http.Response: A reference to the raw HTTP response received from Kraken API.
//   - error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
//
// # Note on error
//
// The error is set only when something wrong has happened either at the HTTP level (while building the request,
// when the server is unreachable, when the API replies with a status code different from 200, ...) , when
// an error happens while parsing the response JSON payload (in that case, error is json.UnmarshalTypeError) or
// when context has expired.
//
// An nil error does not mean everything is OK: You also have to check the response error field for specific
// errors from Kraken API.
//
// # Note on the http.Response
//
// A reference to the received http.Response is always returned but it may be nil if no response was received.
// Some endpoints of the Kraken API include tracing metadata in the response headers. The reference can be used
// to extract the metadata (or any other kind of data that are not used by the API client directly).
//
// Please note response body will always be closed except for RetrieveDataExport.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) WithdrawFunds(ctx context.Context, nonce int64, params funding.WithdrawFundsRequestParameters, opts *funding.WithdrawFundsRequestOptions, secopts *common.SecurityOptions) (*funding.WithdrawFundsResponse, *http.Response, error)

// Trace GetStatusOfRecentWithdrawals - Retrieve information about recent withdrawals. Results are
// sorted by recency, use the cursor parameter to iterate through list of withdrawals (page
// size equal to value of limit) from newest to oldest.
//
// Please note pagination is not used as documentation is unclear about the response format
// in this case.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - opts: GetStatusOfRecentWithdrawals request options. A nil value triggers all default behaviors.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - GetStatusOfRecentWithdrawalsResponse: The parsed response from Kraken API.
//   - http.Response: A reference to the raw HTTP response received from Kraken API.
//   - error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
//
// # Note on error
//
// The error is set only when something wrong has happened either at the HTTP level (while building the request,
// when the server is unreachable, when the API replies with a status code different from 200, ...) , when
// an error happens while parsing the response JSON payload (in that case, error is json.UnmarshalTypeError) or
// when context has expired.
//
// An nil error does not mean everything is OK: You also have to check the response error field for specific
// errors from Kraken API.
//
// # Note on the http.Response
//
// A reference to the received http.Response is always returned but it may be nil if no response was received.
// Some endpoints of the Kraken API include tracing metadata in the response headers. The reference can be used
// to extract the metadata (or any other kind of data that are not used by the API client directly).
//
// Please note response body will always be closed except for RetrieveDataExport.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetStatusOfRecentWithdrawals(ctx context.Context, nonce int64, opts *funding.GetStatusOfRecentWithdrawalsRequestOptions, secopts *common.SecurityOptions) (*funding.GetStatusOfRecentWithdrawalsResponse, *http.Response, error)

// Trace RequestWithdrawalCancellation - Cancel a recently requested withdrawal, if it has not already been successfully processed.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - params: RequestWithdrawalCancellation request parameters.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - RequestWithdrawalCancellationResponse: The parsed response from Kraken API.
//   - http.Response: A reference to the raw HTTP response received from Kraken API.
//   - error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
//
// # Note on error
//
// The error is set only when something wrong has happened either at the HTTP level (while building the request,
// when the server is unreachable, when the API replies with a status code different from 200, ...) , when
// an error happens while parsing the response JSON payload (in that case, error is json.UnmarshalTypeError) or
// when context has expired.
//
// An nil error does not mean everything is OK: You also have to check the response error field for specific
// errors from Kraken API.
//
// # Note on the http.Response
//
// A reference to the received http.Response is always returned but it may be nil if no response was received.
// Some endpoints of the Kraken API include tracing metadata in the response headers. The reference can be used
// to extract the metadata (or any other kind of data that are not used by the API client directly).
//
// Please note response body will always be closed except for RetrieveDataExport.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) RequestWithdrawalCancellation(ctx context.Context, nonce int64, params funding.RequestWithdrawalCancellationRequestParameters, secopts *common.SecurityOptions) (*funding.RequestWithdrawalCancellationResponse, *http.Response, error)

// Trace RequestWalletTransfer - Transfer from a Kraken spot wallet to a Kraken Futures wallet. Note that a
// transfer in the other direction must be requested via the Kraken Futures API endpoint for
// withdrawals to Spot wallets.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - params: RequestWalletTransfer request parameters.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - RequestWalletTransferResponse: The parsed response from Kraken API.
//   - http.Response: A reference to the raw HTTP response received from Kraken API.
//   - error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
//
// # Note on error
//
// The error is set only when something wrong has happened either at the HTTP level (while building the request,
// when the server is unreachable, when the API replies with a status code different from 200, ...) , when
// an error happens while parsing the response JSON payload (in that case, error is json.UnmarshalTypeError) or
// when context has expired.
//
// An nil error does not mean everything is OK: You also have to check the response error field for specific
// errors from Kraken API.
//
// # Note on the http.Response
//
// A reference to the received http.Response is always returned but it may be nil if no response was received.
// Some endpoints of the Kraken API include tracing metadata in the response headers. The reference can be used
// to extract the metadata (or any other kind of data that are not used by the API client directly).
//
// Please note response body will always be closed except for RetrieveDataExport.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) RequestWalletTransfer(ctx context.Context, nonce int64, params funding.RequestWalletTransferRequestParameters, secopts *common.SecurityOptions) (*funding.RequestWalletTransferResponse, *http.Response, error)

// Trace AllocateEarnFunds - Allocate funds to the Strategy.
//
// # Usage tips
//
// This method is asynchronous. A couple of preflight checks are performed synchronously on
// behalf of the method before it is dispatched further. The client is required to poll the
// result using GetAllocationStatus.
//
// There can be only one (de)allocation request in progress for given user and strategy at any
// time. While the operation is in progress:
//
//   - pending attribute in /Earn/Allocations response for the strategy indicates that funds are being allocated.
//   - pending attribute in /Earn/AllocateStatus response will be true.
//
// Following specific errors within Earnings class can be returned by this method:
//
//   - Minimum allocation: EEarnings:Below min:(De)allocation operation amount less than minimum
//   - Allocation in progress: EEarnings:Busy:Another (de)allocation for the same strategy is in progress
//   - Service temporarily unavailable: EEarnings:Busy. Try again in a few minutes.
//   - User tier verification: EEarnings:Permission denied:The user's tier is not high enough
//   - Strategy not found: EGeneral:Invalid arguments:Invalid strategy ID
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - params: AllocateEarnFunds request parameters.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - AllocateEarnFundsResponse: The parsed response from Kraken API.
//   - http.Response: A reference to the raw HTTP response received from Kraken API.
//   - error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
//
// # Note on error
//
// The error is set only when something wrong has happened either at the HTTP level (while building the request,
// when the server is unreachable, when the API replies with a status code different from 200, ...) , when
// an error happens while parsing the response JSON payload (in that case, error is json.UnmarshalTypeError) or
// when context has expired.
//
// An nil error does not mean everything is OK: You also have to check the response error field for specific
// errors from Kraken API.
//
// # Note on the http.Response
//
// A reference to the received http.Response is always returned but it may be nil if no response was received.
// Some endpoints of the Kraken API include tracing metadata in the response headers. The reference can be used
// to extract the metadata (or any other kind of data that are not used by the API client directly).
//
// Please note response body will always be closed except for RetrieveDataExport.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) AllocateEarnFunds(ctx context.Context, nonce int64, params earn.AllocateFundsRequestParameters, secopts *common.SecurityOptions) (*earn.AllocateFundsResponse, *http.Response, error)

// Trace DeallocateEarnFunds - Deallocate funds to the Strategy.
//
// # Usage tips
//
// This method is asynchronous. A couple of preflight checks are performed synchronously on
// behalf of the method before it is dispatched further. The client is required to poll the
// result using GetDeallocationStatus.
//
// There can be only one (de)allocation request in progress for given user and strategy at any
// time. While the operation is in progress:
//
//   - pending attribute in Allocations response for the strategy will hold the amount that is being deallocated (negative amount).
//   - pending attribute in DeallocateStatus response will be true.
//
// Following specific errors within Earnings class can be returned by this method:
//
//   - Minimum allocation: EEarnings:Below min:(De)allocation operation amount less than minimum
//   - Allocation in progress: EEarnings:Busy:Another (de)allocation for the same strategy is in progress
//   - Service temporarily unavailable: EEarnings:Busy. Try again in a few minutes.
//   - Strategy not found: EGeneral:Invalid arguments:Invalid strategy ID
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - params: DeallocateEarnFunds request parameters.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - DeallocateEarnFundsResponse: The parsed response from Kraken API.
//   - http.Response: A reference to the raw HTTP response received from Kraken API.
//   - error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
//
// # Note on error
//
// The error is set only when something wrong has happened either at the HTTP level (while building the request,
// when the server is unreachable, when the API replies with a status code different from 200, ...) , when
// an error happens while parsing the response JSON payload (in that case, error is json.UnmarshalTypeError) or
// when context has expired.
//
// An nil error does not mean everything is OK: You also have to check the response error field for specific
// errors from Kraken API.
//
// # Note on the http.Response
//
// A reference to the received http.Response is always returned but it may be nil if no response was received.
// Some endpoints of the Kraken API include tracing metadata in the response headers. The reference can be used
// to extract the metadata (or any other kind of data that are not used by the API client directly).
//
// Please note response body will always be closed except for RetrieveDataExport.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) DeallocateEarnFunds(ctx context.Context, nonce int64, params earn.DeallocateFundsRequestParameters, secopts *common.SecurityOptions) (*earn.DeallocateFundsResponse, *http.Response, error)

// Trace GetAllocationStatus - Get the status of the last allocation request.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - params: GetAllocationStatus request parameters.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - GetAllocationStatusResponse: The parsed response from Kraken API.
//   - http.Response: A reference to the raw HTTP response received from Kraken API.
//   - error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
//
// # Note on error
//
// The error is set only when something wrong has happened either at the HTTP level (while building the request,
// when the server is unreachable, when the API replies with a status code different from 200, ...) , when
// an error happens while parsing the response JSON payload (in that case, error is json.UnmarshalTypeError) or
// when context has expired.
//
// An nil error does not mean everything is OK: You also have to check the response error field for specific
// errors from Kraken API.
//
// # Note on the http.Response
//
// A reference to the received http.Response is always returned but it may be nil if no response was received.
// Some endpoints of the Kraken API include tracing metadata in the response headers. The reference can be used
// to extract the metadata (or any other kind of data that are not used by the API client directly).
//
// Please note response body will always be closed except for RetrieveDataExport.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetAllocationStatus(ctx context.Context, nonce int64, params earn.GetAllocationStatusRequestParameters, secopts *common.SecurityOptions) (*earn.GetAllocationStatusResponse, *http.Response, error)

// Trace GetDeallocationStatus - Get the status of the last deallocation request.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - params: GetDeallocationStatus request parameters.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - GetDeallocationStatusResponse: The parsed response from Kraken API.
//   - http.Response: A reference to the raw HTTP response received from Kraken API.
//   - error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
//
// # Note on error
//
// The error is set only when something wrong has happened either at the HTTP level (while building the request,
// when the server is unreachable, when the API replies with a status code different from 200, ...) , when
// an error happens while parsing the response JSON payload (in that case, error is json.UnmarshalTypeError) or
// when context has expired.
//
// An nil error does not mean everything is OK: You also have to check the response error field for specific
// errors from Kraken API.
//
// # Note on the http.Response
//
// A reference to the received http.Response is always returned but it may be nil if no response was received.
// Some endpoints of the Kraken API include tracing metadata in the response headers. The reference can be used
// to extract the metadata (or any other kind of data that are not used by the API client directly).
//
// Please note response body will always be closed except for RetrieveDataExport.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetDeallocationStatus(ctx context.Context, nonce int64, params earn.GetDeallocationStatusRequestParameters, secopts *common.SecurityOptions) (*earn.GetDeallocationStatusResponse, *http.Response, error)

// Trace ListEarnStrategies - List earn strategies along with their parameters.
//
// # Usage tips
//
// Returns only strategies that are available to the user based on geographic region.
//
// When the user does not meet the tier restriction, can_allocate will be false and
// allocation_restriction_info indicates Tier as the restriction reason. Earn products
// generally require Intermediate tier. Get your account verified to access earn.
//
// Paging isn't yet implemented, so it the endpoint always returns all data in the first page.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - opts: ListEarnStrategies request options. A nil value triggers all default behaviors.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - ListEarnStrategiesResponse: The parsed response from Kraken API.
//   - http.Response: A reference to the raw HTTP response received from Kraken API.
//   - error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
//
// # Note on error
//
// The error is set only when something wrong has happened either at the HTTP level (while building the request,
// when the server is unreachable, when the API replies with a status code different from 200, ...) , when
// an error happens while parsing the response JSON payload (in that case, error is json.UnmarshalTypeError) or
// when context has expired.
//
// An nil error does not mean everything is OK: You also have to check the response error field for specific
// errors from Kraken API.
//
// # Note on the http.Response
//
// A reference to the received http.Response is always returned but it may be nil if no response was received.
// Some endpoints of the Kraken API include tracing metadata in the response headers. The reference can be used
// to extract the metadata (or any other kind of data that are not used by the API client directly).
//
// Please note response body will always be closed except for RetrieveDataExport.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) ListEarnStrategies(ctx context.Context, nonce int64, opts *earn.ListEarnStrategiesRequestOptions, secopts *common.SecurityOptions) (*earn.ListEarnStrategiesResponse, *http.Response, error)

// Trace ListEarnAllocations - List all allocations for the user.
//
// # Usage tips
//
// By default all allocations are returned, even for strategies that have been used in the past
// and have zero balance now. That is so that the user can see how much was earned with given
// strategy in the past. hide_zero_allocations parameter can be used to remove zero balance
// entries from the output. Paging hasn't been implemented for this method as we don't expect
// the result for a particular user to be overwhelmingly large.
//
// All amounts in the output can be denominated in a currency of user's choice (the converted_asset parameter).
//
// Information about when the next reward will be paid to the client is also provided in the output.
//
// Allocated funds can be in up to 4 states:
//
//   - bonding
//   - allocated
//   - exit_queue (ETH only)
//   - unbonding
//
// Any funds in total not in bonding/unbonding are simply allocated and earning rewards.
// Depending on the strategy funds in the other 3 states can also be earning rewards. Consult
// the output of /Earn/Strategies to know whether bonding/unbonding earn rewards. ETH in
// exit_queue still earns rewards.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - opts: ListEarnAllocations request options. A nil value triggers all default behaviors.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// Note that for ETH, when the funds are in the exit_queue state, the expires time given is the
// time when the funds will have finished unbonding, not when they go from exit queue to unbonding.
//
// (Un)bonding time estimate can be inaccurate right after having (de)allocated the funds. Wait
// 1-2 minutes after (de)allocating to get an accurate result.
//
// # Returns
//
//   - ListEarnAllocationsResponse: The parsed response from Kraken API.
//   - http.Response: A reference to the raw HTTP response received from Kraken API.
//   - error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
//
// # Note on error
//
// The error is set only when something wrong has happened either at the HTTP level (while building the request,
// when the server is unreachable, when the API replies with a status code different from 200, ...) , when
// an error happens while parsing the response JSON payload (in that case, error is json.UnmarshalTypeError) or
// when context has expired.
//
// An nil error does not mean everything is OK: You also have to check the response error field for specific
// errors from Kraken API.
//
// # Note on the http.Response
//
// A reference to the received http.Response is always returned but it may be nil if no response was received.
// Some endpoints of the Kraken API include tracing metadata in the response headers. The reference can be used
// to extract the metadata (or any other kind of data that are not used by the API client directly).
//
// Please note response body will always be closed except for RetrieveDataExport.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) ListEarnAllocations(ctx context.Context, nonce int64, opts *earn.ListEarnAllocationsRequestOptions, secopts *common.SecurityOptions) (*earn.ListEarnAllocationsResponse, *http.Response, error)

// Trace GetWebsocketToken - An authentication token must be requested via this REST API endpoint in
// order to connect to and authenticate with our Websockets API. The token should be used
// within 15 minutes of creation, but it does not expire once a successful Websockets
// connection and private subscription has been made and is maintained.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - GetWebsocketTokenResponse: The parsed response from Kraken API.
//   - http.Response: A reference to the raw HTTP response received from Kraken API.
//   - error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
//
// # Note on error
//
// The error is set only when something wrong has happened either at the HTTP level (while building the request,
// when the server is unreachable, when the API replies with a status code different from 200, ...) , when
// an error happens while parsing the response JSON payload (in that case, error is json.UnmarshalTypeError) or
// when context has expired.
//
// An nil error does not mean everything is OK: You also have to check the response error field for specific
// errors from Kraken API.
//
// # Note on the http.Response
//
// A reference to the received http.Response is always returned but it may be nil if no response was received.
// Some endpoints of the Kraken API include tracing metadata in the response headers. The reference can be used
// to extract the metadata (or any other kind of data that are not used by the API client directly).
//
// Please note response body will always be closed except for RetrieveDataExport.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetWebsocketToken(ctx context.Context, nonce int64, secopts *common.SecurityOptions) (*websocket.GetWebsocketTokenResponse, *http.Response, error)
